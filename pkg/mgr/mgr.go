package mgr

import (
	"fmt"
	"net"
	"tealfs/pkg/cmds"
	"tealfs/pkg/conns"
	"tealfs/pkg/node"
	"tealfs/pkg/proto"
	"tealfs/pkg/raw_net"
	"tealfs/pkg/tnet"
)

type Mgr struct {
	node     node.Node
	userCmds chan cmds.User
	tNet     tnet.TNet
	conns    *conns.Conns
}

func New(userCmds chan cmds.User, tNet tnet.TNet) Mgr {
	base := node.Node{Id: node.NewNodeId(), Address: node.NewAddress(tNet.GetBinding())}
	return Mgr{
		node:     base,
		userCmds: userCmds,
		conns:    conns.New(tNet),
		tNet:     tNet,
	}
}

func (m *Mgr) Start() {
	go m.handleUiCommands()
	go m.acceptConnections()
	go m.readPayloads()
}

func (m *Mgr) Close() {
	m.tNet.Close()
}

func (m *Mgr) GetId() node.Id {
	return m.node.Id
}

func (m *Mgr) acceptConnections() {
	for {
		go m.handleConnection(m.tNet.Accept())
	}
}

func (m *Mgr) readPayloads() {
	for {
		id, payload := m.conns.ReceivePayload()

		switch p := payload.(type) {
		case *proto.SyncNodes:
			fmt.Println(id, p)
		}
	}
}

func (n *Mgr) handleConnection(conn net.Conn) {
	payload := receivePayload(conn)
	switch p := payload.(type) {
	case *proto.Hello:
		n.sendHello(conn)
		node := node.Node{Id: p.NodeId, Address: node.NewAddress(conn.RemoteAddr().String())}
		n.conns.Add(node.Id, node.Address)
	default:
		conn.Close()
	}
}

func (n *Mgr) addRemoteNode(cmd cmds.User) {
	remoteAddress := node.NewAddress(cmd.Argument)
	conn := n.tNet.Dial(remoteAddress.Value)

	n.sendHello(conn)
	payload := receivePayload(conn)

	switch p := payload.(type) {
	case *proto.Hello:
		pld := &proto.Hello{NodeId: p.NodeId}
		n.conns.SendPayload(p.NodeId, pld)
		n.sendHello(conn)
		node := node.Node{Id: p.NodeId, Address: remoteAddress}
		n.conns.Add(node.Id, node.Address)
	default:
		conn.Close()
	}
}

func (n *Mgr) sendHello(conn net.Conn) {
	hello := proto.Hello{NodeId: n.GetId()}
	raw_net.SendPayload(conn, hello.ToBytes())
}

func receivePayload(conn net.Conn) proto.Payload {
	bytes, _ := raw_net.ReadPayload(conn)
	return proto.ToPayload(bytes)
}

func (n *Mgr) handleUiCommands() {
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

func (n *Mgr) addStorage(cmd cmds.User) {
	fmt.Println("Received command: add-storage, location:" + cmd.Argument)
}
