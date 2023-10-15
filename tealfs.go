package main

import (
	"fmt"
	"tealfs/pkg/cmds"
	"tealfs/pkg/node"
	"tealfs/pkg/tnet"
	"tealfs/pkg/ui"
)

func main() {
	userCommands := make(chan cmds.User)
	tNet := tnet.NewTcpNet()

	localNode := node.New(userCommands, tNet)
	localUi := ui.NewUi(&localNode, userCommands)

	localUi.Start()
	localNode.Start()
	fmt.Println("Node: " + localNode.GetAddress())
	select {}
}
