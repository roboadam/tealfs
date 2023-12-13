package mgr

import (
	"fmt"
	"tealfs/pkg/model/events"
	"tealfs/pkg/model/node"
	"tealfs/pkg/proto"
	"tealfs/pkg/tnet"
)

type Mgr struct {
	node     node.Node
	userCmds chan events.Ui
	tNet     tnet.TNet
	conns    *tnet.Conns
}

func New(userCmds chan events.Ui, tNet tnet.TNet) Mgr {
	id := node.NewNodeId()
	n := node.Node{Id: id, Address: node.NewAddress(tNet.GetBinding())}
	return Mgr{
		node:     n,
		userCmds: userCmds,
		conns:    tnet.NewConns(tNet, id),
		tNet:     tNet,
	}
}

func (m *Mgr) Start() {
	go m.handleUiCommands()
	go m.readPayloads()
	go m.handleNewlyConnectdNodes()
}

func (m *Mgr) Close() {
	m.tNet.Close()
}

func (m *Mgr) GetId() node.Id {
	return m.node.Id
}

func (m *Mgr) handleNewlyConnectdNodes() {
	for {
		m.conns.AddedNode()
		m.syncNodes()
	}
}

func (m *Mgr) readPayloads() {
	for {
		remoteId, payload := m.conns.ReceivePayload()

		switch p := payload.(type) {
		case *proto.SyncNodes:
			fmt.Println("readPayloads SyncNodes")
			missingConns := findMyMissingConns(*m.conns, p)
			for _, c := range missingConns.GetValues() {
				m.conns.Add(c)
			}
			if remoteIsMissingNodes(*m.conns, p) {
				toSend := m.BuildSyncNodesPayload()
				m.conns.SendPayload(remoteId, &toSend)
			}
		default:
			fmt.Println("readPayloads default case ")
		}
	}
}

func (m *Mgr) BuildSyncNodesPayload() proto.SyncNodes {
	myNodes := m.conns.GetNodes()
	myNodes.Add(m.node)
	toSend := proto.SyncNodes{Nodes: myNodes}
	return toSend
}

func (m *Mgr) addRemoteNode(cmd events.Ui) {
	remoteAddress := node.NewAddress(cmd.Argument)
	m.conns.Add(tnet.NewConn(remoteAddress))
	m.syncNodes()
}

func (m *Mgr) handleUiCommands() {
	for {
		command := <-m.userCmds
		switch command.EventType {
		case events.ConnectTo:
			m.addRemoteNode(command)
		case events.AddStorage:
			m.addStorage(command)
		}
	}
}

func (m *Mgr) addStorage(cmd events.Ui) {
	fmt.Println("Received command: add-storage, location:" + cmd.Argument)
}

func (m *Mgr) syncNodes() {
	allIds := m.conns.GetIds()
	for _, id := range allIds.GetValues() {
		payload := m.BuildSyncNodesPayload()
		fmt.Println("mgr.syncNodes to " + id.String())
		m.conns.SendPayload(id, &payload)
	}
}
