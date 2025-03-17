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
	"fmt"
	"os"
	"strconv"
	"strings"
	"tealfs/pkg/conns"
	"tealfs/pkg/disk"
	"tealfs/pkg/mgr"
	"tealfs/pkg/model"
	"tealfs/pkg/ui"
	"tealfs/pkg/webdav"

	log "github.com/sirupsen/logrus"
)

func main() {
	log.SetLevel(log.TraceLevel)
	if len(os.Args) < 7 {
		fmt.Fprintln(os.Stderr, os.Args[0], "<config path> <disk paths> <webdav address> <ui address> <node address> <free bytes>")
		os.Exit(1)
	}

	val, err := strconv.ParseUint(os.Args[6], 10, 32)
	if err != nil {
		fmt.Fprintln(os.Stderr, os.Args[0], "<config path> <disk paths> <webdav address> <ui address> <node address> <free bytes>")
		os.Exit(1)
	}

	freeBytes := uint32(val)
	disks := strings.Split(os.Args[2], ",")

	_ = startTealFs(os.Args[1], disks, os.Args[3], os.Args[4], os.Args[5], freeBytes, context.Background())
}

func startTealFs(globalPath string, disks []string, webdavAddress string, uiAddress string, nodeAddress string, freeBytes uint32, ctx context.Context) error {
	chansize := 0
	m := mgr.NewWithChanSize(chansize, nodeAddress, globalPath, &disk.DiskFileOps{}, model.Mirrored, freeBytes, disks)
	_ = conns.NewConns(
		m.ConnsMgrStatuses,
		m.ConnsMgrReceives,
		m.MgrConnsConnectTos,
		m.MgrConnsSends,
		&conns.TcpConnectionProvider{},
		nodeAddress,
		m.NodeId,
		ctx,
	)
	for _, diskId := range m.DiskIds {
		writeChan := m.MgrDiskWrites[diskId]
		readChan := m.MgrDiskReads[diskId]
		p := disk.NewPath(m.Disks()[diskId], &disk.DiskFileOps{})
		_ = disk.New(p,
			model.NewNodeId(),
			writeChan,
			readChan,
			m.DiskMgrWrites,
			m.DiskMgrReads,
			ctx,
		)
	}
	_ = ui.NewUi(m.UiMgrConnectTos, m.MgrUiStatuses, &ui.HttpHtmlOps{}, uiAddress, ctx)
	_ = webdav.New(
		m.NodeId,
		m.WebdavMgrGets,
		m.WebdavMgrPuts,
		m.WebdavMgrBroadcast,
		m.MgrWebdavGets,
		m.MgrWebdavPuts,
		m.MgrWebdavBroadcast,
		webdavAddress,
		ctx,
		&disk.DiskFileOps{},
		globalPath,
		chansize,
	)
	err := m.Start(ctx)
	if err != nil {
		return err
	}
	<-ctx.Done()
	return nil
}
