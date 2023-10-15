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
	node := Node{
		Id:          NewNodeId(),
		userCmds:    userCmds,
		remoteNodes: NewRemoteNodes(),
		tNet:        tNet,
	}

	go node.handleUiCommands()
	go node.acceptConnections()

	return node
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

func (node *Node) handleUiCommands() {
	for {
		command := <-node.userCmds
		switch command.CmdType {
		case cmds.ConnectTo:
			node.addConnection(command)
		case cmds.AddStorage:
			node.addStorage(command)
		}
	}
}

func (node *Node) addConnection(cmd cmds.User) {

	remoteNode := NewRemoteNode(node.Id, cmd.Argument, node.tNet)

	node.remoteNodes.Add(*remoteNode)
	fmt.Println("Received command: add-connection, address:" + cmd.Argument + ", added connection id:" + remoteNode.NodeId.value.String())
}

func (node *Node) addStorage(cmd cmds.User) {
	fmt.Println("Received command: add-storage, location:" + cmd.Argument)
}
