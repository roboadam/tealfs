package main

import (
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
	if len(os.Args) < 4 {
		fmt.Fprintln(os.Stderr, os.Args[0], "<storage path> <webdav address> <ui address>")
		os.Exit(1)
	}
	m := mgr.NewWithChanSize(1)
	_ = conns.NewConns(
		m.ConnsMgrStatuses,
		m.ConnsMgrReceives,
		m.MgrConnsConnectTos,
		m.MgrConnsSends,
		&conns.TcpConnectionProvider{},
	)
	p := disk.NewPath(os.Args[1], &disk.DiskFileOps{})
	_ = disk.New(p,
		model.NewNodeId(),
		m.MgrDiskWrites,
		m.MgrDiskReads,
		m.DiskMgrWrites,
		m.DiskMgrReads,
	)
	_ = ui.NewUi(m.UiMgrConnectTos, m.MgrUiStatuses, &ui.HttpHtmlOps{}, os.Args[3])
	_ = webdav.New(
		m.NodeId,
		m.WebdavMgrGets,
		m.WebdavMgrPuts,
		m.MgrWebdavGets,
		m.MgrWebdavPuts,
		os.Args[2],
	)
	m.Start()
	select {}
}
