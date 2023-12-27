package mgr

import (
	"fmt"
	"tealfs/pkg/model/events"
	"tealfs/pkg/model/node"
	"tealfs/pkg/proto"
	"tealfs/pkg/store"
	"tealfs/pkg/tnet"
	"tealfs/pkg/util"
)

type Mgr struct {
	node     node.Node
	userCmds chan events.Event
	tNet     tnet.TNet
	conns    *tnet.Conns
	store    *store.Paths
}

func New(userCmds chan events.Event, tNet tnet.TNet) Mgr {
	id := node.NewNodeId()
	fmt.Printf("New Node Id %s\n", id.String())
	n := node.Node{Id: id, Address: node.NewAddress(tNet.GetBinding())}
	conns := tnet.NewConns(tNet, id)
	return Mgr{
		node:     n,
		userCmds: userCmds,
		conns:    conns,
		tNet:     tNet,
	}
}

func (m *Mgr) Start() {
	go m.handleUiCommands()
	go m.readPayloads()
	go m.handleNewlyConnectedNodes()
}

func (m *Mgr) Close() {
	m.tNet.Close()
}

func (m *Mgr) GetId() node.Id {
	return m.node.Id
}

func (m *Mgr) handleNewlyConnectedNodes() {
	for {
		_ = m.conns.AddedNode()
		m.syncNodes()
	}
}

func (m *Mgr) readPayloads() {
	for {
		remoteId, payload := m.conns.ReceivePayload()

		switch p := payload.(type) {
		case *proto.SyncNodes:
			missingConns := findMyMissingConns(*m.conns, p)
			for _, c := range missingConns.GetValues() {
				m.conns.Add(c)
			}
			if remoteIsMissingNodes(*m.conns, p) {
				toSend := m.BuildSyncNodesPayload()
				m.conns.SendPayload(remoteId, &toSend)
			}
		default:
			// Do nothing
		}
	}
}

func (m *Mgr) BuildSyncNodesPayload() proto.SyncNodes {
	myNodes := m.conns.GetNodes()
	myNodes.Add(m.node)
	toSend := proto.SyncNodes{Nodes: myNodes}
	return toSend
}

func (m *Mgr) GetRemoteNodes() util.Set[node.Node] {
	result := m.conns.GetNodes()
	return result
}

func (m *Mgr) addRemoteNode(cmd events.Event) {
	remoteAddress := node.NewAddress(cmd.GetString())
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
		case events.AddData:
			m.addData(command)
		}
	}
}

func (m *Mgr) addData(_ events.Event) {

}

func (m *Mgr) addStorage(event events.Event) {

}

func (m *Mgr) syncNodes() {
	allIds := m.conns.GetIds()
	for _, id := range allIds.GetValues() {
		payload := m.BuildSyncNodesPayload()
		m.conns.SendPayload(id, &payload)
	}
}
