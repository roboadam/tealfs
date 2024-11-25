// Copyright (C) 2024 Adam Hess
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
	"tealfs/pkg/conns"
	"tealfs/pkg/disk"
	"tealfs/pkg/mgr"
	"tealfs/pkg/model"
	"tealfs/pkg/ui"
	"tealfs/pkg/webdav"
)

func main() {
	if len(os.Args) < 5 {
		fmt.Fprintln(os.Stderr, os.Args[0], "<storage path> <webdav address> <ui address> <node address>")
		os.Exit(1)
	}
	startTealFs(os.Args[1], os.Args[2], os.Args[3], os.Args[4], context.Background())
}

func startTealFs(storagePath string, webdavAddress string, uiAddress string, nodeAddress string, ctx context.Context) {
	m := mgr.NewWithChanSize(1, nodeAddress)
	_ = conns.NewConns(
		m.ConnsMgrStatuses,
		m.ConnsMgrReceives,
		m.MgrConnsConnectTos,
		m.MgrConnsSends,
		&conns.TcpConnectionProvider{},
		nodeAddress,
		m.NodeId,
	)
	p := disk.NewPath(storagePath, &disk.DiskFileOps{})
	_ = disk.New(p,
		model.NewNodeId(),
		m.MgrDiskWrites,
		m.MgrDiskReads,
		m.DiskMgrWrites,
		m.DiskMgrReads,
	)
	_ = ui.NewUi(m.UiMgrConnectTos, m.MgrUiStatuses, &ui.HttpHtmlOps{}, uiAddress)
	_ = webdav.New(
		m.NodeId,
		m.WebdavMgrGets,
		m.WebdavMgrPuts,
		m.MgrWebdavGets,
		m.MgrWebdavPuts,
		webdavAddress,
		ctx,
	)
	m.Start()
	<-ctx.Done()
}
