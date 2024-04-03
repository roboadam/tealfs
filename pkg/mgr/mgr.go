package mgr

import (
	"tealfs/pkg/nodes"
	"tealfs/pkg/proto"
	"tealfs/pkg/set"
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
	nodeConnMap NodeConnMap
	nodeId      nodes.Id
	connAddress map[ConnId]string
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
		nodeConnMap:               NodeConnMap{},
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
		connId, success := m.nodeConnMap.Conn(node)
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
		m.nodes.Add(p.NodeId)
		m.nodeConnMap.Add(p.NodeId, i.ConnId)

		syncNodes := m.syncNodesPayloadToSend()
		m.MgrConnsSends <- MgrConnsSend{
			ConnId:  i.ConnId,
			Payload: &syncNodes,
		}
	case *proto.SyncNodes:
		remoteNodes := p.GetNodes()
		missing := remoteNodes.Minus(&m.nodes)
		missing.Remove(m.nodeId)
		// TODO Need to add address to Node or whatever goes over the wire to include the address to connect
	}
}
func (m *Mgr) handleReads(i DiskMgrRead) {}

type IAmReq struct {
	nodeId nodes.Id
	connId ConnId
	resp   chan<- IAmResp
}
type IAmResp struct {
}

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

type MgrDiskSave struct{}
type DiskMgrRead struct{}
type MgrDiskRead struct{}

func (m *Mgr) addNodeToCluster(r IAmReq) {
	m.nodes.Add(r.nodeId)
	m.nodeConnMap.Add(r.nodeId, r.connId)
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
