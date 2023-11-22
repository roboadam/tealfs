package manager

import (
	"fmt"
	"net"
	"tealfs/pkg/cmds"
	"tealfs/pkg/conns"
	"tealfs/pkg/node"
	"tealfs/pkg/raw_net"
	"tealfs/pkg/tnet"
)

type Manager struct {
	node     node.Node
	userCmds chan cmds.User
	tNet     tnet.TNet
	conns    *conns.Conns
}

func New(userCmds chan cmds.User, tNet tnet.TNet) Manager {
	base := node.Node{Id: node.NewNodeId(), Address: node.NewAddress(tNet.GetBinding())}
	return Manager{
		node:     base,
		userCmds: userCmds,
		conns:    conns.New(tNet),
		tNet:     tNet,
	}
}

func (n *Manager) Start() {
	go n.handleUiCommands()
	go n.acceptConnections()
	go n.readPayloads()
}

func (n *Manager) Close() {
	n.tNet.Close()
}

func (n *Manager) GetId() node.Id {
	return n.node.Id
}

func (n *Manager) acceptConnections() {
	for {
		go n.handleConnection(n.tNet.Accept())
	}
}

func (n *Manager) readPayloads() {
	// for {
	// 	id, payload := n.remoteNodes.ReceivePayload()
	// 	switch p:= payload.(type) {
	// 	case *SyncNodes:

	// 	}
	// }
}

func (n *Manager) handleConnection(conn net.Conn) {
	payload := receivePayload(conn)
	switch p := payload.(type) {
	case *node.Hello:
		n.sendHello(conn)
		node := node.Node{Id: p.NodeId, Address: node.NewAddress(conn.RemoteAddr().String())}
		n.conns.Add(node.Id, node.Address)
	default:
		conn.Close()
	}
}

func (n *Manager) addRemoteNode(cmd cmds.User) {
	remoteAddress := node.NewAddress(cmd.Argument)
	conn := n.tNet.Dial(remoteAddress.Value)

	n.sendHello(conn)
	payload := receivePayload(conn)

	switch p := payload.(type) {
	case *node.Hello:
		n.sendHello(conn)
		node := node.Node{Id: p.NodeId, Address: remoteAddress}
		n.conns.Add(node.Id, node.Address)
	default:
		conn.Close()
	}
}

func (n *Manager) sendHello(conn net.Conn) {
	hello := node.Hello{NodeId: n.GetId()}
	raw_net.SendPayload(conn, hello.ToBytes())
}

func receivePayload(conn net.Conn) node.Payload {
	bytes, _ := raw_net.ReadPayload(conn)
	return node.ToPayload(bytes)
}

func (n *Manager) handleUiCommands() {
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

func (n *Manager) addStorage(cmd cmds.User) {
	fmt.Println("Received command: add-storage, location:" + cmd.Argument)
}

// func (n *Manager) GetRemoteNode(id Id) (*Node, error) {
// 	return n.conns.GetNode(id)
// }
