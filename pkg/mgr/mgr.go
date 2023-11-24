package mgr

import (
	"fmt"
	"tealfs/pkg/cmds"
	"tealfs/pkg/conns"
	"tealfs/pkg/node"
	"tealfs/pkg/proto"
	"tealfs/pkg/tnet"
)

type Mgr struct {
	node     node.Node
	userCmds chan cmds.User
	tNet     tnet.TNet
	conns    *conns.Conns
}

func New(userCmds chan cmds.User, tNet tnet.TNet) Mgr {
	myNodeId := node.NewNodeId()
	base := node.Node{Id: myNodeId, Address: node.NewAddress(tNet.GetBinding())}
	return Mgr{
		node:     base,
		userCmds: userCmds,
		conns:    conns.New(tNet, myNodeId),
		tNet:     tNet,
	}
}

func (m *Mgr) Start() {
	go m.handleUiCommands()
	go m.readPayloads()
}

func (m *Mgr) Close() {
	m.tNet.Close()
}

func (m *Mgr) GetId() node.Id {
	return m.node.Id
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

func (n *Mgr) addRemoteNode(cmd cmds.User) {
	remoteAddress := node.NewAddress(cmd.Argument)
	n.conns.Add(remoteAddress)

	// n.sendHello(conn)
	// payload := receivePayload(conn)

	// switch p := payload.(type) {
	// case *proto.Hello:
	// 	pld := &proto.Hello{NodeId: p.NodeId}
	// 	n.conns.SendPayload(p.NodeId, pld)
	// 	n.sendHello(conn)
	// 	node := node.Node{Id: p.NodeId, Address: remoteAddress}
	// 	n.conns.Add(node.Id, node.Address)
	// default:
	// 	conn.Close()
	// }
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
