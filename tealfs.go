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
	if len(os.Args) < 5 {
		fmt.Fprintln(os.Stderr, os.Args[0], "<storage path> <webdav address> <ui address> <node address>")
		os.Exit(1)
	}
	m := mgr.NewWithChanSize(1, os.Args[4])
	_ = conns.NewConns(
		m.ConnsMgrStatuses,
		m.ConnsMgrReceives,
		m.MgrConnsConnectTos,
		m.MgrConnsSends,
		&conns.TcpConnectionProvider{},
		os.Args[4],
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
