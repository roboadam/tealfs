package main

import (
	"tealfs/pkg/mgr"
	"tealfs/pkg/model/cmds"
	"tealfs/pkg/tnet"
	"tealfs/pkg/ui"
)

func main() {
	userCommands := make(chan cmds.User)
	tNet := tnet.NewTcpNet("127.0.0.1:0")

	localNode := mgr.New(userCommands, tNet)
	localUi := ui.NewUi(&localNode, userCommands)

	localUi.Start()
	localNode.Start()
	select {}
}
