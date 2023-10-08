package main

import (
	"fmt"
	"tealfs/pkg/cmds"
	"tealfs/pkg/node"
	"tealfs/pkg/ui"
)

func main() {
	userCommands := make(chan cmds.User)

	localNode := node.New(userCommands)
	localUi := ui.NewUi(&localNode, userCommands)

	go localUi.Start()
	localNode.Listen()
	fmt.Println("Node: " + localNode.GetAddress())
	select {}
}
