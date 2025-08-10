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

	if len(os.Args) < 4 {
		fmt.Fprintln(os.Stderr, os.Args[0], "<webdav address> <ui address> <node address>")
		os.Exit(1)
	}

	_ = startTealFs(configDir, os.Args[1], os.Args[2], os.Args[3], context.Background())
}

func startTealFs(globalPath string, webdavAddress string, uiAddress string, nodeAddress string, ctx context.Context) error {
	log.SetLevel(log.DebugLevel)
	chansize := 0
	connReqs := make(chan model.ConnectToNodeReq)
	nodeConnMapper := model.NewNodeConnectionMapper()
	m := mgr.New(chansize, nodeAddress, globalPath, &disk.DiskFileOps{}, nodeConnMapper, ctx)
	m.ConnectToNodeReqs = connReqs

	custodianCommands := make(chan custodian.Command, chansize)
	m.CustodianCommands = custodianCommands
	custodian := custodian.NewCustodian(chansize)
	custodian.Commands = custodianCommands
	custodian.Start(ctx)

	iamConnIds := make(chan conns.IamConnId, 1)
	incomingSyncNodes := make(chan model.SyncNodes, 1)
	sendSyncNodes := make(chan struct{}, 1)
	saveCluster := make(chan struct{}, 1)

	connsMain := conns.NewConns(
		m.ConnsMgrStatuses,
		m.ConnsMgrReceives,
		connReqs,
		m.MgrConnsSends,
		&conns.TcpConnectionProvider{},
		nodeAddress,
		m.NodeId,
		ctx,
	)

	connsMain.OutIamConnId = iamConnIds
	connsMain.OutSyncNodes = incomingSyncNodes

	connsIamReceiver := conns.IamReceiver{
		InIam:            iamConnIds,
		OutSendSyncNodes: sendSyncNodes,
		OutSaveCluster:   saveCluster,
		Mapper:           nodeConnMapper,
	}
	go connsIamReceiver.Start(ctx)

	clusterSaver := conns.ClusterSaver{
		Save:           saveCluster,
		NodeConnMapper: nodeConnMapper,
		SavePath:       globalPath,
		FileOps:        &disk.DiskFileOps{},
	}
	go clusterSaver.Start(ctx)

	receiveSyncNodes := conns.ReceiveSyncNodes{
		InSyncNodes:    incomingSyncNodes,
		OutConnectTo:   connReqs,
		NodeConnMapper: nodeConnMapper,
		NodeId:         m.NodeId,
	}
	go receiveSyncNodes.Start(ctx)

	sendSyncNodesProc := conns.SendSyncNodes{
		InSendSyncNodes: sendSyncNodes,
		OutSendPayloads: m.MgrConnsSends,
		NodeConnMapper:  nodeConnMapper,
	}
	go sendSyncNodesProc.Start(ctx)

	clusterLoader := conns.ClusterLoader{
		NodeConnMapper: nodeConnMapper,
		FileOps:        &disk.DiskFileOps{},
		SavePath:       globalPath,
	}
	go clusterLoader.Load(ctx)

	newAddDiskReqs := make(chan model.AddDiskReq)
	_ = ui.NewUi(
		connReqs,
		m.MgrUiConnectionStatuses,
		newAddDiskReqs,
		m.MgrUiDiskStatuses,
		&ui.HttpHtmlOps{},
		m.NodeId,
		uiAddress,
		ctx,
	)

	addedDisksSaver := make(chan *disk.Disk)
	addedDisksReader := make(chan *disk.Disk)

	localAddDiskReqs := make(chan model.AddDiskReq)
	remoteAddDiskReqs := make(chan model.AddDiskReq)

	connsMain.OutAddDiskReq = newAddDiskReqs
	disks := disk.NewDisks(m.NodeId)
	disks.InAddDiskReq = newAddDiskReqs
	disks.OutLocalAddDiskReq = localAddDiskReqs
	disks.OutRemoteAddDiskReq = remoteAddDiskReqs
	m.AllDiskIds = &disks.AllDiskIds
	go disks.Start(ctx)

	diskLoader := disk.DiskLoader{
		FileOps:    &disk.DiskFileOps{},
		SavePath:   globalPath,
		OutAddDisk: newAddDiskReqs,
	}
	go diskLoader.LoadDisks(ctx)

	iamDiskUpdates := make(chan []model.AddDiskReq, 1)
	saveDisks := make(chan struct{}, 1)

	diskSaver := disk.DiskSaver{
		FileOps:    &disk.DiskFileOps{},
		LoadPath:   globalPath,
		AllDiskIds: &disks.AllDiskIds,
		Save:       saveDisks,
	}
	go diskSaver.Start(ctx)

	localDiskAdder := disk.LocalDiskAdder{
		InAddDiskReq: localAddDiskReqs,
		OutAddLocalDisk: []chan<- *disk.Disk{
			addedDisksSaver,
			addedDisksReader,
		},
		OutIamDiskUpdate: iamDiskUpdates,
		OutSave:          saveDisks,
		FileOps:          &disk.DiskFileOps{},
		Disks:            &disks.Disks,
		Distributer:      &disks.Distributer,
		AllDiskIds:       &disks.AllDiskIds,
	}
	go localDiskAdder.Start(ctx)

	iamSender := disk.IamSender{
		InIamDiskUpdate: iamDiskUpdates,
		OutSends:        m.MgrConnsSends,
		Mapper:          nodeConnMapper,
		NodeId:          m.NodeId,
		Address:         webdavAddress,
	}
	go iamSender.Start(ctx)

	iams := make(chan model.IAm)
	connsMain.OutIam = iams
	iamReceiver := disk.IamReceiver{
		InIam:       iams,
		OutSave:     saveDisks,
		Distributer: &disks.Distributer,
		AllDiskIds:  &disks.AllDiskIds,
	}
	go iamReceiver.Start(ctx)

	webdavPutReq := make(chan model.PutBlockReq)
	webdavPutResp := make(chan model.PutBlockResp)

	localSave := make(chan blocksaver.SaveToDiskReq)
	remoteSave := make(chan blocksaver.SaveToDiskReq)
	saveResp := make(chan blocksaver.SaveToDiskResp)

	connsMain.OutSaveToDiskReq = localSave
	connsMain.OutSaveToDiskResp = saveResp

	bs := blocksaver.BlockSaver{
		Req:         webdavPutReq,
		RemoteDest:  remoteSave,
		LocalDest:   localSave,
		InResp:      saveResp,
		Resp:        webdavPutResp,
		Distributer: &disks.Distributer,
		NodeId:      m.NodeId,
	}
	go bs.Start(ctx)

	lbs := blocksaver.LocalBlockSaver{
		Req:   localSave,
		Disks: &disks.Disks,
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

	connsMain.OutGetFromDiskReq = localReadDest
	connsMain.OutGetFromDiskResp = readResp

	br := blockreader.BlockReader{
		Req:         webdavGetReq,
		RemoteDest:  remoteReadDest,
		LocalDest:   localReadDest,
		InResp:      readResp,
		Resp:        webdavGetResp,
		Distributer: &disks.Distributer,
		NodeId:      m.NodeId,
	}
	go br.Start(ctx)

	lbr := blockreader.LocalBlockReader{
		Req:   localReadDest,
		Disks: &disks.Disks,
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
