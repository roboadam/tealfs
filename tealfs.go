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
	"tealfs/pkg/conns"
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

	log.SetLevel(log.TraceLevel)
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
	chansize := 0
	m := mgr.NewWithChanSize(chansize, nodeAddress, globalPath, &disk.DiskFileOps{}, model.Mirrored, freeBytes)
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
		writeChan := m.MgrDiskWrites[diskId.Id]
		readChan := m.MgrDiskReads[diskId.Id]
		p := disk.NewPath(diskId.Path, &disk.DiskFileOps{})
		_ = disk.New(p,
			model.NewNodeId(),
			writeChan,
			readChan,
			m.DiskMgrWrites,
			m.DiskMgrReads,
			ctx,
		)
	}
	_ = ui.NewUi(m.UiMgrConnectTos, m.MgrUiConnectionStatuses, &ui.HttpHtmlOps{}, uiAddress, ctx)
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
