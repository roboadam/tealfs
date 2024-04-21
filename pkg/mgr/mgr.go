package mgr

import (
	"tealfs/pkg/hash"
	"tealfs/pkg/nodes"
	"tealfs/pkg/proto"
	"tealfs/pkg/set"
	"tealfs/pkg/store"
	"tealfs/pkg/store/dist"
)

type Mgr struct {
	UiMgrConnectTos    chan UiMgrConnectTo
	ConnsMgrStatuses   chan ConnsMgrStatus
	ConnsMgrReceives   chan ConnsMgrReceive
	DiskMgrReads       chan ReadResult
	DiskMgrWrites      chan WriteResult
	WebdavMgrGets      chan store.Id
	WebdavMgrPuts      chan store.Block
	MgrConnsConnectTos chan MgrConnsConnectTo
	MgrConnsSends      chan MgrConnsSend
	MgrDiskWrites      chan store.Block
	MgrDiskReads       chan store.Id
	MgrWebdavGets      chan ReadResult
	MgrWebdavPuts      chan WriteResult

	nodes       set.Set[nodes.Id]
	nodeConnMap set.Bimap[nodes.Id, ConnId]
	nodeId      nodes.Id
	connAddress map[ConnId]string
	distributer dist.Distributer
}

func NewNew() Mgr {
	mgr := Mgr{
		UiMgrConnectTos:    make(chan UiMgrConnectTo, 1),
		ConnsMgrStatuses:   make(chan ConnsMgrStatus, 1),
		ConnsMgrReceives:   make(chan ConnsMgrReceive, 1),
		DiskMgrReads:       make(chan ReadResult, 1),
		WebdavMgrGets:      make(chan store.Id, 1),
		WebdavMgrPuts:      make(chan store.Block, 1),
		MgrConnsConnectTos: make(chan MgrConnsConnectTo, 1),
		MgrConnsSends:      make(chan MgrConnsSend, 1),
		MgrDiskWrites:      make(chan store.Block, 1),
		MgrDiskReads:       make(chan store.Id, 1),
		MgrWebdavGets:      make(chan ReadResult, 1),
		MgrWebdavPuts:      make(chan WriteResult, 1),
		nodes:              set.NewSet[nodes.Id](),
		nodeId:             nodes.NewNodeId(),
	}

	return mgr
}

func (m *Mgr) Start() {
	go m.eventLoop()
}

func (m *Mgr) eventLoop() {
	for {
		select {
		case r := <-m.UiMgrConnectTos:
			m.handleConnectToReq(r)
		case r := <-m.ConnsMgrStatuses:
			m.handleConnectedStatus(r)
		case r := <-m.ConnsMgrReceives:
			m.handleReceives(r)
		case r := <-m.DiskMgrReads:
			m.handleReads(r)
		case r := <-m.WebdavMgrGets:
			m.handleGets(r)
		case r := <-m.WebdavMgrPuts:
			m.handlePuts(r)
		}
	}
}

func (m *Mgr) handleConnectToReq(i UiMgrConnectTo) {
	m.MgrConnsConnectTos <- MgrConnsConnectTo{Address: i.Address}
}

func (m *Mgr) syncNodesPayloadToSend() proto.SyncNodes {
	result := proto.SyncNodes{}
	for _, node := range m.nodes.GetValues() {
		connId, success := m.nodeConnMap.Get1(node)
		if success {
			if address, ok := m.connAddress[connId]; ok {
				result.Nodes.Add(struct {
					Node    nodes.Id
					Address string
				}{Node: node, Address: address})
			}
		}
	}
	return result
}

func (m *Mgr) handleReceives(i ConnsMgrReceive) {
	switch p := i.Payload.(type) {
	case *proto.IAm:
		m.addNodeToCluster(p.NodeId, i.ConnId)
		syncNodes := m.syncNodesPayloadToSend()
		for _, n := range m.nodes.GetValues() {
			connId, ok := m.nodeConnMap.Get1(n)
			if ok {
				m.MgrConnsSends <- MgrConnsSend{
					ConnId:  connId,
					Payload: &syncNodes,
				}
			}
		}
	case *proto.SyncNodes:
		remoteNodes := p.GetNodes()
		localNodes := m.nodes.Clone()
		localNodes.Add(m.nodeId)
		missing := remoteNodes.Minus(&m.nodes)
		for _, n := range missing.GetValues() {
			address := p.AddressForNode(n)
			m.MgrConnsConnectTos <- MgrConnsConnectTo{Address: address}
		}
	case *proto.SaveData:
		n := m.distributer.NodeIdForStoreId(p.Block.Id)
		if m.nodeId == n {
			m.MgrDiskWrites <- p.Block
		} else {
			c, ok := m.nodeConnMap.Get1(n)
			if ok {
				m.MgrConnsSends <- MgrConnsSend{ConnId: c, Payload: p}
			} else {
				m.MgrDiskWrites <- p.Block
			}
		}
	}
}

func (m *Mgr) handleReads(i ReadResult) {
	// Todo: need to handle read results from disk
}

type MgrConnsConnectTo struct {
	Address string
}

type ConnsMgrStatus struct {
	Type          ConnectedStatus
	RemoteAddress string
	Msg           string
	Id            ConnId
}
type ConnectedStatus int

const (
	Connected ConnectedStatus = iota
	NotConnected
)

type MgrConnsSend struct {
	ConnId  ConnId
	Payload proto.Payload
}

type ConnsMgrReceive struct {
	ConnId  ConnId
	Payload proto.Payload
}

type MgrDiskSave struct {
	Hash hash.Hash
	Data []byte
}
type ReadResult struct {
	Ok      bool
	Message string
	Block   store.Block
}
type WriteResult struct {
	Ok      bool
	Message string
}

func (m *Mgr) addNodeToCluster(n nodes.Id, c ConnId) {
	m.nodes.Add(n)
	m.nodeConnMap.Add(n, c)
	m.distributer.SetWeight(n, 1)
}

func (m *Mgr) handleConnectedStatus(cs ConnsMgrStatus) {
	switch cs.Type {
	case Connected:
		m.connAddress[cs.Id] = cs.RemoteAddress
		m.MgrConnsSends <- MgrConnsSend{
			ConnId:  cs.Id,
			Payload: &proto.IAm{NodeId: m.nodeId},
		}
	case NotConnected:
		println("Not Connected")
	}
}

func (m *Mgr) handleGets(r store.Id) {
	n := m.distributer.NodeIdForStoreId(r)
	if m.nodeId == n {
		m.MgrDiskReads <- r
	} else {
		m.MgrConnsSends <- proto.
	}
	// Todo: Take get request from webdav and request data from another server or get it from the local disk
}

func (m *Mgr) handlePuts(_ store.Block) {
	// Todo: Take put request from webdav and save the data to another server or save it to the local disk
}
