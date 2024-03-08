package main

import (
	"os"
	"strconv"
	"tealfs/pkg/mgr"
	"tealfs/pkg/model/events"
	"tealfs/pkg/store"
	"tealfs/pkg/tnet"
)

func main() {
	userCommands := make(chan events.Event)

	tNet := tnet.NewTcpNet("127.0.0.1:" + strconv.Itoa(nodePort()))

	localNode := mgr.New(userCommands, tNet, store.NewPath(os.Args[0]))
	//localUi := ui.NewUi(&localNode, userCommands)
	//
	//localUi.Start()
	localNode.Start()
	select {}
}

func nodePort() int {
	if len(os.Args) > 0 {
		num, err := strconv.Atoi(os.Args[0])
		if err == nil {
			return num
		}
	}
	return 0
}
