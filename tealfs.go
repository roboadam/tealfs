package main

import (
	"fmt"
	"tealfs/pkg/cmds"
	"tealfs/pkg/node"
	"tealfs/pkg/ui"
)

func main() {
	userCommands := make(chan cmds.User)

	localNode := node.NewNode(userCommands)
	localUi := ui.NewUi(&localNode, userCommands)

	go localUi.Start()
	localNode.Listen()
	fmt.Println("Node: " + localNode.GetAddress().String())
	select {}
}
