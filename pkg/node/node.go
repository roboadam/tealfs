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
	tnet        tnet.TNet
	connections *RemoteNodes
	HostToBind  string
}

func New(userCmds chan cmds.User, tnet tnet.TNet) Node {
	node := Node{
		Id:          NewNodeId(),
		userCmds:    userCmds,
		connections: NewRemoteNodes(),
		tnet:        tnet,
		HostToBind:  "",
	}

	go node.handleUiCommands()
	go node.acceptConnections()

	return node
}

func (node *Node) GetAddress() string {
	return node.tnet.GetAddress()
}

func (node *Node) Close() {
	node.tnet.Close()
}

func (node *Node) acceptConnections() {
	for {
		go node.handleConnection(node.tnet.Accept())
	}
}

func (node *Node) Listen() {
	node.tnet.Listen(node.HostToBind + ":0")
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
	conn := RemoteNode{
		NodeId:  node.Id,
		Address: cmd.Argument,
		Conn:    nil,
	}

	node.connections.AddConnection(conn)
	fmt.Println("Received command: add-connection, address:" + cmd.Argument + ", added connection id:" + conn.NodeId.value.String())
}

func (node *Node) addStorage(cmd cmds.User) {
	fmt.Println("Received command: add-storage, location:" + cmd.Argument)
}
