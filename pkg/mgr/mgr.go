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
	DiskMgrReads       chan proto.ReadResult
	DiskMgrWrites      chan WriteResult
	WebdavMgrGets      chan ReadRequest
	WebdavMgrPuts      chan store.Block
	MgrConnsConnectTos chan MgrConnsConnectTo
	MgrConnsSends      chan MgrConnsSend
	MgrDiskWrites      chan store.Block
	MgrDiskReads       chan ReadRequest
	MgrWebdavGets      chan proto.ReadResult
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
		DiskMgrReads:       make(chan proto.ReadResult, 1),
		WebdavMgrGets:      make(chan ReadRequest, 1),
		WebdavMgrPuts:      make(chan store.Block, 1),
		MgrConnsConnectTos: make(chan MgrConnsConnectTo, 1),
		MgrConnsSends:      make(chan MgrConnsSend, 1),
		MgrDiskWrites:      make(chan store.Block, 1),
		MgrDiskReads:       make(chan ReadRequest, 1),
		MgrWebdavGets:      make(chan proto.ReadResult, 1),
		MgrWebdavPuts:      make(chan WriteResult, 1),
		nodes:              set.NewSet[nodes.Id](),
		nodeId:             nodes.NewNodeId(),
		connAddress:        make(map[ConnId]string),
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
			m.handleDiskReads(r)
		case r := <-m.WebdavMgrGets:
			m.handleWebdavGets(r)
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

func (m *Mgr) handleDiskReads(r proto.ReadResult) {
	if r.Caller == m.nodeId {
		m.MgrWebdavGets <- r
	} else {
		c, ok := m.nodeConnMap.Get1(r.Caller)
		if ok {
			m.MgrConnsSends <- MgrConnsSend{ConnId: c, Payload: &r}
		} else {
			// Todo: need a ticket to create queuing for offline nodes
		}
	}
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
type ReadRequest struct {
	Caller  nodes.Id
	BlockId store.Id
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

func (m *Mgr) handleWebdavGets(rr ReadRequest) {
	n := m.distributer.NodeIdForStoreId(rr.BlockId)
	if m.nodeId == n {
		m.MgrDiskReads <- rr
	} else {
		c, ok := m.nodeConnMap.Get1(n)
		if ok {
			m.MgrConnsSends <- MgrConnsSend{
				ConnId:  c,
				Payload: &proto.ReadRequest{Caller: rr.Caller, BlockId: rr.BlockId},
			}
		} else {
			m.MgrWebdavGets <- proto.ReadResult{
				Ok:      false,
				Message: "Not connected",
				Block:   store.Block{Id: rr.BlockId},
				Caller:  rr.Caller,
			}
		}
	}
}

func (m *Mgr) handlePuts(_ store.Block) {
	// Todo: Take put request from webdav and save the data to another server or save it to the local disk
}
