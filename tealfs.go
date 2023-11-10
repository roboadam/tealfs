package main

import (
	"tealfs/pkg/cmds"
	"tealfs/pkg/node"
	"tealfs/pkg/tnet"
	"tealfs/pkg/ui"
)

func main() {
	userCommands := make(chan cmds.User)
	tNet := tnet.NewTcpNet("127.0.0.1:0")

	localNode := node.New(userCommands, tNet)
	localUi := ui.NewUi(&localNode, userCommands)

	localUi.Start()
	localNode.Start()
	select {}
}
