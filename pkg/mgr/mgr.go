package mgr

import (
	"tealfs/pkg/hash"
	"tealfs/pkg/model/events"
	"tealfs/pkg/model/node"
	"tealfs/pkg/proto"
	"tealfs/pkg/store"
	d "tealfs/pkg/store/dist"
	"tealfs/pkg/tnet"
	"tealfs/pkg/util"
)

type Mgr struct {
	node   node.Node
	events chan events.Event
	tNet   tnet.TNet
	conns  *tnet.Conns
	store  *store.Store
	dist   *d.Distributer
}

func New(events chan events.Event, tNet tnet.TNet, path store.Path) Mgr {
	id := node.NewNodeId()
	n := node.Node{Id: id, Address: node.NewAddress(tNet.GetBinding())}
	println("node ", id.String(), "has address", tNet.GetBinding())
	conns := tnet.NewConns(tNet, id)
	s := store.New(path, id)
	dist := d.NewDistributer()
	dist.SetWeight(id, 1)
	return Mgr{
		node:   n,
		events: events,
		conns:  conns,
		tNet:   tNet,
		store:  &s,
		dist:   dist,
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

func (m *Mgr) PrintDist() {
	m.dist.PrintDist()
}

func (m *Mgr) GetId() node.Id {
	return m.node.Id
}

func (m *Mgr) handleNewlyConnectedNodes() {
	for {
		n := m.conns.AddedNode()
		m.dist.SetWeight(n, 1)
		m.syncNodes()
	}
}

func (m *Mgr) readPayloads() {
	for {
		println("trace readpayloads 1", m.GetId().String())
		remoteId, payload := m.conns.ReceivePayload()
		println("trace readpayloads 2", m.GetId().String())

		switch p := payload.(type) {
		case *proto.SyncNodes:
			println("trace readpayloads 3", m.GetId().String())
			missingConns := findMyMissingConns(*m.conns, p)
			println("trace readpayloads 3.1", m.GetId().String())
			for _, c := range missingConns.GetValues() {
				println("trace readpayloads 3.2", m.GetId().String())
				m.conns.Add(c)
				println("trace readpayloads 3.3", m.GetId().String())
			}
			println("trace readpayloads 3.4", m.GetId().String())
			if remoteIsMissingNodes(*m.conns, p) {
				println("trace readpayloads 3.5", m.GetId().String())
				toSend := m.buildSyncNodesPayload()
				println("trace readpayloads 3.6", m.GetId().String())
				m.conns.SendPayload(remoteId, &toSend)
				println("trace readpayloads 3.7", m.GetId().String())
			}
			println("trace readpayloads 4", m.GetId().String())
		case *proto.SaveData:
			println("I got a savedata in mgr!")
			println("trace readpayloads 5", m.GetId().String())
			m.saveToAppropriateNode(p.Data)
			println("trace readpayloads 6", m.GetId().String())
		default:
			println("trace readpayloads 7", m.GetId().String())
			// Do nothing
		}
	}
}

func (m *Mgr) buildSyncNodesPayload() proto.SyncNodes {
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
		command := <-m.events
		switch command.EventType {
		case events.ConnectTo:
			m.addRemoteNode(command)
		case events.AddData:
			m.addData(command)
		case events.ReadData:
			m.readData(command)
		}
	}
}

func (m *Mgr) addData(d events.Event) {
	data := d.GetBytes()
	m.saveToAppropriateNode(data)
}

func (m *Mgr) saveToAppropriateNode(data []byte) {
	h := hash.ForData(data)
	id := m.dist.NodeIdForHash(h)
	if m.GetId() == id {
		println("saveToAppropriateNode:forme")
		m.store.Save(h, data)
	} else {
		println("saveToAppropriateNode:foryou")
		payload := proto.ToSaveData(data)
		m.conns.SendPayload(id, payload)
	}
}

func (m *Mgr) readData(d events.Event) {
	h := hash.FromRaw(d.GetBytes())
	r := d.GetResult()
	r <- m.store.Read(h)
}

func (m *Mgr) syncNodes() {
	allIds := m.conns.GetIds()
	for _, id := range allIds.GetValues() {
		payload := m.buildSyncNodesPayload()
		m.conns.SendPayload(id, &payload)
	}
}
