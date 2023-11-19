package node

import (
	"fmt"
	"net"
	"tealfs/pkg/cmds"
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
	base := Node{Id: NewNodeId(), Address: NewAddress(tNet.GetBinding())}
	return LocalNode{
		base:        base,
		userCmds:    userCmds,
		remoteNodes: NewRemoteNodes(base),
		tNet:        tNet,
	}
}

func (n *LocalNode) Start() {
	go n.handleUiCommands()
	go n.acceptConnections()
	go n.readPayloads()
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

func (n *LocalNode) readPayloads() {
	for {
		id, payload := n.remoteNodes.ReceivePayload()
		switch p:= payload.(type) {
		case *SyncNodes:
			
		}
	}
}

func (n *LocalNode) handleConnection(conn net.Conn) {
	payload := receivePayload(conn)
	switch p := payload.(type) {
	case *Hello:
		n.sendHello(conn)
		node := Node{Id: p.NodeId, Address: NewAddress(conn.RemoteAddr().String())}
		n.remoteNodes.Add(node, conn)
	default:
		conn.Close()
	}
}

func (n *LocalNode) addRemoteNode(cmd cmds.User) {
	remoteAddress := NewAddress(cmd.Argument)
	conn := n.tNet.Dial(remoteAddress.value)

	n.sendHello(conn)
	payload := receivePayload(conn)

	switch p := payload.(type) {
	case *Hello:
		n.sendHello(conn)
		node := Node{Id: p.NodeId, Address: remoteAddress}
		n.remoteNodes.Add(node, conn)
	default:
		conn.Close()
	}
}

func (n *LocalNode) sendHello(conn net.Conn) {
	hello := Hello{NodeId: n.GetId()}
	raw_net.SendPayload(conn, hello.ToBytes())
}

func receivePayload(conn net.Conn) Payload {
	bytes, _ := raw_net.ReadPayload(conn)
	return ToPayload(bytes)
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

func (n *LocalNode) GetRemoteNode(id Id) (*Node, error) {
	return n.remoteNodes.GetNode(id)
}
