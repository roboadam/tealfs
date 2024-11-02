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
	if len(os.Args) < 2 {
		fmt.Fprintln(os.Stderr, "Specify the path to store data")
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
	_ = ui.NewUi(m.UiMgrConnectTos, m.ConnsMgrStatuses, &ui.HttpHtmlOps{})
	_ = webdav.New(m.NodeId, &webdav.HttpWebdavOps{})
	select {}
}
