// Copyright (C) 2025 Adam Hess
//
// This program is free software: you can redistribute it and/or modify it under
// the terms of the GNU Affero General Public License as published by the Free
// Software Foundation, version 3.
//
// This program is distributed in the hope that it will be useful, but WITHOUT
// ANY WARRANTY; without even the implied warranty of MERCHANTABILITY or
// FITNESS FOR A PARTICULAR PURPOSE. See the GNU Affero General Public License
// for more details.
//
// You should have received a copy of the GNU Affero General Public License
// along with this program. If not, see <https://www.gnu.org/licenses/>.

package main

import (
	"context"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strconv"
	"tealfs/pkg/blockreader"
	"tealfs/pkg/blocksaver"
	"tealfs/pkg/conns"
	"tealfs/pkg/custodian"
	"tealfs/pkg/disk"
	"tealfs/pkg/mgr"
	"tealfs/pkg/model"
	"tealfs/pkg/ui"
	"tealfs/pkg/webdav"

	log "github.com/sirupsen/logrus"
)

func main() {
	configDir, err := os.UserConfigDir()
	if err != nil {
		fmt.Println("Error getting user config directory:", err)
		return
	}
	configDir = filepath.Join(configDir, "tealfs")
	if err = os.Mkdir(configDir, 0700); err != nil && !errors.Is(err, fs.ErrExist) {
		fmt.Printf("unable to create config directory: {%s}. error: %s\n", configDir, err)
		os.Exit(1)
	}

	if len(os.Args) < 5 {
		fmt.Fprintln(os.Stderr, os.Args[0], "<webdav address> <ui address> <node address> <free bytes>")
		os.Exit(1)
	}

	val, err := strconv.ParseUint(os.Args[4], 10, 32)
	if err != nil {
		fmt.Fprintln(os.Stderr, os.Args[0], "<webdav address> <ui address> <node address> <free bytes>")
		os.Exit(1)
	}

	freeBytes := uint32(val)

	_ = startTealFs(configDir, os.Args[1], os.Args[2], os.Args[3], freeBytes, context.Background())
}

func startTealFs(globalPath string, webdavAddress string, uiAddress string, nodeAddress string, freeBytes uint32, ctx context.Context) error {
	log.SetLevel(log.DebugLevel)
	chansize := 0
	connReqs := make(chan model.ConnectToNodeReq)
	nodeConnMapper := model.NewNodeConnectionMapper()
	m := mgr.New(chansize, nodeAddress, globalPath, &disk.DiskFileOps{}, freeBytes, nodeConnMapper, ctx)
	m.ConnectToNodeReqs = connReqs

	custodianCommands := make(chan custodian.Command, chansize)
	m.CustodianCommands = custodianCommands
	custodian := custodian.NewCustodian(chansize)
	custodian.Commands = custodianCommands
	custodian.Start(ctx)

	conns := conns.NewConns(
		m.ConnsMgrStatuses,
		m.ConnsMgrReceives,
		connReqs,
		m.MgrConnsSends,
		&conns.TcpConnectionProvider{},
		nodeAddress,
		m.NodeId,
		ctx,
	)
	_ = ui.NewUi(
		connReqs,
		m.MgrUiConnectionStatuses,
		m.UiMgrDisk,
		m.MgrUiDiskStatuses,
		&ui.HttpHtmlOps{},
		m.NodeId,
		uiAddress,
		ctx,
	)

	addedDisksSaver := make(chan *disk.Disk)
	addedDisksReader := make(chan *disk.Disk)
	m.AddedDisk = []chan<- *disk.Disk{
		addedDisksSaver,
		addedDisksReader,
	}

	webdavPutReq := make(chan model.PutBlockReq)
	webdavPutResp := make(chan model.PutBlockResp)

	localSave := make(chan blocksaver.SaveToDiskReq)
	remoteSave := make(chan blocksaver.SaveToDiskReq)
	saveResp := make(chan blocksaver.SaveToDiskResp)

	conns.OutSaveToDiskReq = localSave
	conns.OutSaveToDiskResp = saveResp

	bs := blocksaver.BlockSaver{
		Req:         webdavPutReq,
		RemoteDest:  remoteSave,
		LocalDest:   localSave,
		InResp:      saveResp,
		Resp:        webdavPutResp,
		Distributer: &m.MirrorDistributer,
		NodeId:      m.NodeId,
	}
	go bs.Start(ctx)

	lbs := blocksaver.LocalBlockSaver{
		Req:   localSave,
		Disks: &m.Disks,
	}
	go lbs.Start(ctx)

	rbs := blocksaver.RemoteBlockSaver{
		Req:         remoteSave,
		Sends:       m.MgrConnsSends,
		NoConnResp:  saveResp,
		NodeConnMap: nodeConnMapper,
	}
	go rbs.Start(ctx)

	lbsr := blocksaver.LocalBlockSaveResponses{
		InDisks:             addedDisksSaver,
		LocalWriteResponses: saveResp,
		Sends:               m.MgrConnsSends,
		NodeConnMap:         nodeConnMapper,
		NodeId:              m.NodeId,
	}
	go lbsr.Start(ctx)

	webdavGetReq := make(chan model.GetBlockReq)
	webdavGetResp := make(chan model.GetBlockResp)
	localReadDest := make(chan blockreader.GetFromDiskReq)
	remoteReadDest := make(chan blockreader.GetFromDiskReq)
	readResp := make(chan blockreader.GetFromDiskResp)

	br := blockreader.BlockReader{
		Req:         webdavGetReq,
		RemoteDest:  remoteReadDest,
		LocalDest:   localReadDest,
		InResp:      readResp,
		Resp:        webdavGetResp,
		Distributer: &m.MirrorDistributer,
		NodeId:      m.NodeId,
	}
	go br.Start(ctx)

	lbr := blockreader.LocalBlockReader{
		Req:   localReadDest,
		Disks: &m.Disks,
	}
	go lbr.Start(ctx)

	rbr := blockreader.RemoteBlockReader{
		Req:         remoteReadDest,
		Sends:       m.MgrConnsSends,
		NoConnResp:  readResp,
		NodeConnMap: nodeConnMapper,
	}
	go rbr.Start(ctx)

	lbrr := blockreader.LocalBlockReadResponses{
		InDisks:            addedDisksReader,
		LocalReadResponses: readResp,
		Sends:              m.MgrConnsSends,
		NodeConnMap:        nodeConnMapper,
		NodeId:             m.NodeId,
	}
	go lbrr.Start(ctx)

	_ = webdav.New(
		m.NodeId,
		webdavGetReq,
		webdavPutReq,
		m.WebdavMgrBroadcast,
		webdavGetResp,
		webdavPutResp,
		m.MgrWebdavBroadcast,
		webdavAddress,
		ctx,
		&disk.DiskFileOps{},
		globalPath,
		chansize,
	)
	m.Start()
	<-ctx.Done()
	return nil
}
