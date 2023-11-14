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
	payload := n.ReceivePayload(conn)
	if payload.Type == proto.Hello() {
		remoteId, _ := HelloFromBytes(payload.Data)
		n.sendHello(conn)
		node := Node{Id: remoteId, Address: NewAddress(conn.RemoteAddr().String())}
		n.remoteNodes.Add(node, conn)
	}
}

func (n *LocalNode) addRemoteNode(cmd cmds.User) {
	remoteAddress := NewAddress(cmd.Argument)
	conn := n.tNet.Dial(remoteAddress.value)

	n.sendHello(conn)
	payload := n.ReceivePayload(conn)
	if payload.Type == proto.Hello() {
		remoteId, _ := HelloFromBytes(payload.Data)
		n.remoteNodes.Add(Node{Id: remoteId, Address: remoteAddress}, conn)
	}
}

func (*LocalNode) ReceivePayload(conn net.Conn) proto.Payload {
	rawData, _ := raw_net.ReadBytes(conn, proto.HeaderLen)
	header, _ := proto.HeaderFromBytes(rawData)
	rawData, _ = raw_net.ReadBytes(conn, header.Len)
	return proto.Payload{Type: header.Typ, Data: rawData}
}

func (n *LocalNode) sendHello(conn net.Conn) {
	payload := HelloToBytes(n.GetId())
	header := proto.Header{Typ: proto.Hello(), Len: uint32(len(payload))}
	raw_net.SendBytes(conn, header.ToBytes())
	raw_net.SendBytes(conn, payload)
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
