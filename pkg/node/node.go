package node

import (
	"fmt"
	"net"
	"tealfs/pkg/cmds"
	"tealfs/pkg/raw_net"
	"tealfs/pkg/tnet"
)

type Node struct {
	Id          Id
	userCmds    chan cmds.User
	tNet        tnet.TNet
	remoteNodes *RemoteNodes
}

func New(userCmds chan cmds.User, tNet tnet.TNet) Node {
	return Node{
		Id:          NewNodeId(),
		userCmds:    userCmds,
		remoteNodes: NewRemoteNodes(),
		tNet:        tNet,
	}
}

func (n *Node) Start() {
	go n.handleUiCommands()
	go n.acceptConnections()
}

func (n *Node) Close() {
	n.tNet.Close()
}

func (n *Node) acceptConnections() {
	for {
		go n.handleConnection(n.tNet.Accept())
	}
}

func (n *Node) handleConnection(conn net.Conn) {
	for {
		intFromConn, _ := raw_net.IntFrom(conn)
		fmt.Println("Received:", intFromConn)
	}
}

func (n *Node) addConnection(cmd cmds.User) {

	remoteNode := NewRemoteNode(n.Id, cmd.Argument, n.tNet)

	n.remoteNodes.Add(*remoteNode)
	fmt.Println("Received command: add-connection, address:" + cmd.Argument + ", added connection id:" + remoteNode.NodeId.value.String())
}

func (n *Node) handleUiCommands() {
	for {
		command := <-n.userCmds
		switch command.CmdType {
		case cmds.ConnectTo:
			n.addConnection(command)
		case cmds.AddStorage:
			n.addStorage(command)
		}
	}
}

func (n *Node) addStorage(cmd cmds.User) {
	fmt.Println("Received command: add-storage, location:" + cmd.Argument)
}
