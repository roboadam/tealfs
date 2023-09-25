package main

import (
	"flag"
	"fmt"
	"math/rand"
	"tealfs/pkg/cmds"
	"tealfs/pkg/node"
	"tealfs/pkg/ui"
)

func main() {
	webPort, nodePort := ports()
	hostId := hostId()
	userCmds := make(chan cmds.User)

	node := node.NewNode(userCmds)

	go ui.StartUi(webPort, nodePort, userCmds, hostId)
	go node.Start()
	fmt.Println("UI: http://localhost:" + fmt.Sprint(webPort))
	fmt.Println("Node: localhost:" + fmt.Sprint(nodePort))
	select {}
}

func ports() (int, int) {
	portPtr := flag.Int("port", 8001, "Port for web UI")
	flag.Parse()

	portWeb := *portPtr
	portNode := portWeb + 1

	return portWeb, portNode
}

func hostId() uint32 {
	return rand.Uint32()
}
