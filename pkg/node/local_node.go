package node

import (
	"fmt"
	"net"
	"tealfs/pkg/cmds"
	"tealfs/pkg/proto"
	"tealfs/pkg/raw_net"
	"tealfs/pkg/tnet"
)

type LocalNode struct {
	base        Node
	userCmds    chan cmds.User
	tNet        tnet.TNet
	remoteNodes *RemoteNodes
}

func New(userCmds chan cmds.User, tNet tnet.TNet) LocalNode {
	return LocalNode{
		base:        Node{Id: NewNodeId(), Address: NewAddress(tNet.GetBinding())},
		userCmds:    userCmds,
		remoteNodes: NewRemoteNodes(),
		tNet:        tNet,
	}
}

func (n *LocalNode) Start() {
	go n.handleUiCommands()
	go n.acceptConnections()
}

func (n *LocalNode) Close() {
	n.tNet.Close()
}

func (n *LocalNode) GetId() Id {
	return n.base.Id
}

func (n *LocalNode) acceptConnections() {
	for {
		go n.handleConnection(n.tNet.Accept())
	}
}

func (n *LocalNode) handleConnection(conn net.Conn) {
	command, length, _ := raw_net.CommandAndLength(conn)
	raw_data, _ := raw_net.ReadBytes(conn, length)
	if command == proto.NetCmd{value:proto.Hello} {
		length, _ := raw_net.UInt32From(conn)
		rawId, _ := raw_net.StringFrom(conn, int(length))
		remoteNode := NewRemoteNode(IdFromRaw(rawId), conn.RemoteAddr().String(), n.tNet)
		n.remoteNodes.Add(*remoteNode)
	}
}

func (n *LocalNode) addRemoteNode(cmd cmds.User) {

	remoteNode := NewRemoteNode(n.GetId(), cmd.Argument, n.tNet)

	n.remoteNodes.Add(*remoteNode)
	remoteNode.Connect()
	remoteNode.SendHello(n.GetId())
	fmt.Println("Received command: add-connection, address:" + cmd.Argument + ", added connection id:" + remoteNode.Id.String())
}

func (n *LocalNode) handleUiCommands() {
	for {
		command := <-n.userCmds
		switch command.CmdType {
		case cmds.ConnectTo:
			n.addRemoteNode(command)
		case cmds.AddStorage:
			n.addStorage(command)
		}
	}
}

func (n *LocalNode) addStorage(cmd cmds.User) {
	fmt.Println("Received command: add-storage, location:" + cmd.Argument)
}

func (n *LocalNode) GetRemoteNode(id Id) (*RemoteNode, error) {
	return n.remoteNodes.GetConnection(id)
}
