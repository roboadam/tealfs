package mgr

import (
	"fmt"
	"tealfs/pkg/model/events"
	"tealfs/pkg/model/node"
	"tealfs/pkg/proto"
	"tealfs/pkg/tnet"
	"tealfs/pkg/util"
)

type Mgr struct {
	node     node.Node
	userCmds chan events.Ui
	tNet     tnet.TNet
	conns    *tnet.Conns
	Debug    bool
}

func NewDebug(userCmds chan events.Ui, tNet tnet.TNet, debug bool) Mgr {
	r := New(userCmds, tNet)
	r.Debug = true
	return r
}

func New(userCmds chan events.Ui, tNet tnet.TNet) Mgr {
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
		m.conns.AddedNode()
		m.syncNodes()
	}
}

func (m *Mgr) readPayloads() {
	for {
		remoteId, payload := m.conns.ReceivePayload()

		switch p := payload.(type) {
		case *proto.SyncNodes:
			if m.Debug {
				fmt.Println("M:readPayloads SyncNodes")
			}
			missingConns := findMyMissingConns(*m.conns, p)
			for _, c := range missingConns.GetValues() {
				if m.Debug {
					fmt.Println("M:Adding missing conn")
				}
				m.conns.Add(c)
			}
			if remoteIsMissingNodes(*m.conns, p) {
				toSend := m.BuildSyncNodesPayload()
				m.conns.SendPayload(remoteId, &toSend)
			}
		default:
			if m.Debug {
				fmt.Println("M:readPayloads default case ")
			}
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
	return m.conns.GetNodes()
}

func (m *Mgr) addRemoteNode(cmd events.Ui) {
	if m.Debug {
		fmt.Println("Connect To Event: " + cmd.Argument)
	}
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
	if m.Debug {
		fmt.Println("M:Received command: add-storage, location:" + cmd.Argument)
	}
}

func (m *Mgr) syncNodes() {
	allIds := m.conns.GetIds()
	for _, id := range allIds.GetValues() {
		payload := m.BuildSyncNodesPayload()
		if m.Debug {
			fmt.Println("M:mgr.syncNodes to " + id.String())
		}
		m.conns.SendPayload(id, &payload)
	}
}
