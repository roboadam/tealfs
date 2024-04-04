package mgr

import (
	"tealfs/pkg/hash"
	"tealfs/pkg/nodes"
	"tealfs/pkg/proto"
	"tealfs/pkg/set"
	"tealfs/pkg/store/dist"
)

type Mgr struct {
	UiMgrConnectTos           chan UiMgrConnectTo
	ConnsMgrConnectedStatuses chan ConnsMgrConnectedStatus
	ConnsMgrReceives          chan ConnsMgrReceive
	DiskMgrRead               chan DiskMgrRead
	MgrConnsConnectTos        chan MgrConnsConnectTo
	MgrConnsSends             chan MgrConnsSend
	MgrDiskSaves              chan MgrDiskSave
	MgrDiskRead               chan MgrDiskRead

	nodes       set.Set[nodes.Id]
	nodeConnMap set.Bimap[nodes.Id, ConnId]
	nodeId      nodes.Id
	connAddress map[ConnId]string
	distributer dist.Distributer
}

func NewNew() Mgr {
	mgr := Mgr{
		UiMgrConnectTos:           make(chan UiMgrConnectTo, 1),
		ConnsMgrConnectedStatuses: make(chan ConnsMgrConnectedStatus, 1),
		ConnsMgrReceives:          make(chan ConnsMgrReceive, 1),
		DiskMgrRead:               make(chan DiskMgrRead, 1),
		MgrConnsConnectTos:        make(chan MgrConnsConnectTo, 1),
		MgrConnsSends:             make(chan MgrConnsSend, 1),
		MgrDiskSaves:              make(chan MgrDiskSave, 1),
		nodes:                     set.NewSet[nodes.Id](),
		nodeId:                    nodes.NewNodeId(),
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
		case r := <-m.ConnsMgrConnectedStatuses:
			m.handleConnectedStatus(r)
		case r := <-m.ConnsMgrReceives:
			m.handleReceives(r)
		case r := <-m.DiskMgrRead:
			m.handleReads(r)
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
		h := hash.ForData(p.Data)
		n := m.distributer.NodeIdForHash(h)
		if m.nodeId == n {
			m.MgrDiskSaves <- MgrDiskSave{Hash: h, Data: p.Data}
		} else {
			c, ok := m.nodeConnMap.Get1(n)
			if ok {
				m.MgrConnsSends <- MgrConnsSend{ConnId: c, Payload: p}
			} else {
				// Todo: not sure what to do here
			}
			// Todo: Otherwise send a SaveData to the destination node
		}
	}
}
func (m *Mgr) handleReads(i DiskMgrRead) {}

type MyNodesReq struct {
	resp chan<- MyNodesResp
}
type MyNodesResp struct {
}

type SaveToClusterReq struct {
	resp chan<- SaveToClusterResp
}
type SaveToClusterResp struct {
}

type SaveToDiskReq struct {
	resp chan<- SaveToDiskResp
}
type SaveToDiskResp struct {
}

type MgrConnsConnectTo struct {
	Address string
}

type ConnsMgrConnectedStatus struct {
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
type DiskMgrRead struct{}
type MgrDiskRead struct{}

func (m *Mgr) addNodeToCluster(n nodes.Id, c ConnId) {
	m.nodes.Add(n)
	m.nodeConnMap.Add(n, c)
	m.distributer.SetWeight(n, 1)
}
func (m *Mgr) handleMyNodes(_ MyNodesReq)             {}
func (m *Mgr) handleSaveToCluster(_ SaveToClusterReq) {}
func (m *Mgr) handleSaveToDisk(_ SaveToDiskReq)       {}

func (m *Mgr) handleConnectedStatus(cs ConnsMgrConnectedStatus) {
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
