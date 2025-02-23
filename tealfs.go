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
	if len(os.Args) < 6 {
		fmt.Fprintln(os.Stderr, os.Args[0], "<storage path> <webdav address> <ui address> <node address> <free bytes>")
		os.Exit(1)
	}

	val, err := strconv.ParseUint(os.Args[5], 10, 32)
	if err != nil {
		fmt.Fprintln(os.Stderr, os.Args[0], "<storage path> <webdav address> <ui address> <node address> <free bytes>")
		os.Exit(1)
	}

	freeBytes := uint32(val)

	_ = startTealFs(os.Args[1], os.Args[2], os.Args[3], os.Args[4], freeBytes, context.Background())
}

func startTealFs(storagePath string, webdavAddress string, uiAddress string, nodeAddress string, freeBytes uint32, ctx context.Context) error {
	m := mgr.NewWithChanSize(2, nodeAddress, storagePath, &disk.DiskFileOps{}, model.Mirrored, freeBytes)
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
	p := disk.NewPath(storagePath, &disk.DiskFileOps{})
	_ = disk.New(p,
		model.NewNodeId(),
		m.MgrDiskWrites,
		m.MgrDiskReads,
		m.DiskMgrWrites,
		m.DiskMgrReads,
	)
	_ = ui.NewUi(m.UiMgrConnectTos, m.MgrUiStatuses, &ui.HttpHtmlOps{}, uiAddress, ctx)
	_ = webdav.New(
		m.NodeId,
		m.WebdavMgrGets,
		m.WebdavMgrPuts,
		m.MgrWebdavGets,
		m.MgrWebdavPuts,
		webdavAddress,
		ctx,
	)
	err := m.Start()
	if err != nil {
		return err
	}
	<-ctx.Done()
	return nil
}
