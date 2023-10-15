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

func (node *Node) Start() {
	go node.handleUiCommands()
	go node.acceptConnections()
}

func (node *Node) Close() {
	node.tNet.Close()
}

func (node *Node) acceptConnections() {
	for {
		go node.handleConnection(node.tNet.Accept())
	}
}

func (node *Node) handleConnection(conn net.Conn) {
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

func (n Node) addStorage(cmd cmds.User) {
	fmt.Println("Received command: add-storage, location:" + cmd.Argument)
}
