package mgr

import (
	"tealfs/pkg/nodes"
	"tealfs/pkg/proto"
	"tealfs/pkg/set"
)

type MgrNew struct {
	UiMgrConnectTos           chan UiMgrConnectTo
	ConnsMgrConnectedStatuses chan ConnsMgrConnectedStatus
	ConnsMgrReceives          chan ConnsMgrReceive
	DiskMgrRead               chan DiskMgrRead
	MgrConnsConnectTos        chan MgrConnsConnectTo
	MgrConnsSends             chan MgrConnsSend
	MgrDiskSaves              chan MgrDiskSave
	MgrDiskRead               chan MgrDiskRead

	nodes       nodes.Nodes
	nodeConnMap NodeConnMap
	nodeId      nodes.Id
}

func NewNew() MgrNew {
	mgr := MgrNew{
		UiMgrConnectTos:           make(chan UiMgrConnectTo, 1),
		ConnsMgrConnectedStatuses: make(chan ConnsMgrConnectedStatus, 1),
		ConnsMgrReceives:          make(chan ConnsMgrReceive, 1),
		DiskMgrRead:               make(chan DiskMgrRead, 1),
		MgrConnsConnectTos:        make(chan MgrConnsConnectTo, 1),
		MgrConnsSends:             make(chan MgrConnsSend, 1),
		MgrDiskSaves:              make(chan MgrDiskSave, 1),
		nodes:                     nodes.Nodes{},
		nodeConnMap:               NodeConnMap{},
		nodeId:                    nodes.NewNodeId(),
	}

	return mgr
}

func (m *MgrNew) Start() {
	go m.eventLoop()
}

func (m *MgrNew) eventLoop() {
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

func (m *MgrNew) handleConnectToReq(i UiMgrConnectTo) {
	m.MgrConnsConnectTos <- MgrConnsConnectTo{Address: i.Address}
}

func (m *MgrNew) handleReceives(i ConnsMgrReceive) {
	switch p := i.Payload.(type) {
	case *proto.IAm:
		m.nodes.AddOrUpdate(nodes.NodeNew{Id: p.NodeId})
		m.nodeConnMap.Add(p.NodeId, i.ConnId)

		syncNodes := proto.SyncNodes{Nodes: set.NewSet[nodes.Node]()}
		m.MgrConnsSends <- MgrConnsSend{
			ConnId:  i.ConnId,
			Payload: &syncNodes,
		}
		// TODO send out sync nodes if anything changed
	}
}
func (m *MgrNew) handleReads(i DiskMgrRead) {}

type IAmReq struct {
	nodeId nodes.Id
	connId ConnNewId
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
	Type ConnectedStatus
	Msg  string
	Id   ConnNewId
}
type ConnectedStatus int

const (
	Connected ConnectedStatus = iota
	NotConnected
)

type MgrConnsSend struct {
	ConnId  ConnNewId
	Payload proto.Payload
}

type ConnsMgrReceive struct {
	ConnId  ConnNewId
	Payload proto.Payload
}

type MgrDiskSave struct{}
type DiskMgrRead struct{}
type MgrDiskRead struct{}

func (m *MgrNew) addNodeToCluster(r IAmReq) {
	m.nodes.AddOrUpdate(nodes.NodeNew{Id: r.nodeId})
	m.nodeConnMap.Add(r.nodeId, r.connId)
}
func (m *MgrNew) handleMyNodes(_ MyNodesReq)             {}
func (m *MgrNew) handleSaveToCluster(_ SaveToClusterReq) {}
func (m *MgrNew) handleSaveToDisk(_ SaveToDiskReq)       {}

func (m *MgrNew) handleConnectedStatus(cs ConnsMgrConnectedStatus) {
	switch cs.Type {
	case Connected:
		m.MgrConnsSends <- MgrConnsSend{
			ConnId:  cs.Id,
			Payload: &proto.IAm{NodeId: m.nodeId},
		}
	case NotConnected:
		println("Not Connected")
	}
}
