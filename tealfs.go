package main

import (
	"flag"
	"fmt"
	"tealfs/pkg/cmds"
	"tealfs/pkg/node"
	"tealfs/pkg/ui"
)

func main() {
	webPort, _ := ports()
	userCmds := make(chan cmds.User)

	node := node.NewNode(userCmds)
	ui := ui.NewUi(&node, userCmds)

	go ui.Start()
	node.Listen()
	fmt.Println("UI: http://localhost:" + fmt.Sprint(webPort))
	fmt.Println("Node: " + node.GetAddress().String())
	select {}
}

func ports() (int, int) {
	portPtr := flag.Int("port", 8001, "Port for web UI")
	flag.Parse()

	portWeb := *portPtr
	portNode := portWeb + 1

	return portWeb, portNode
}
