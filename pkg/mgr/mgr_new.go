package mgr

import (
	"tealfs/pkg/nodes"
)

type MgrNew struct {
	InUiConnectTo          chan InUiConnectTo
	OutConnsConnectTo      chan OutConnsConnectTo
	InConnsConnectedStatus chan InConnsConnectedStatus
	OutConnsSend           chan OutConnsSend
	InConnsSendStatus      chan InConnsSendStatus
	InConnsReceive         chan InConnsReceive
	OutDiskSave            chan OutDiskSave
	OutDiskSaveStatus      chan OutDiskSaveStatus

	nodes       nodes.Nodes
	nodeConnMap NodeConnMap
	nodeId      nodes.Id
}

func NewNew() MgrNew {
	mgr := MgrNew{
		InUiConnectTo:          make(chan InUiConnectTo, 1),
		OutConnsConnectTo:      make(chan OutConnsConnectTo, 1),
		InConnsConnectedStatus: make(chan InConnsConnectedStatus, 1),
		OutConnsSend:           make(chan OutConnsSend, 1),
		InConnsSendStatus:      make(chan InConnsSendStatus, 1),
		InConnsReceive:         make(chan InConnsReceive, 1),
		OutDiskSave:            make(chan OutDiskSave, 1),
		OutDiskSaveStatus:      make(chan OutDiskSaveStatus, 1),
		nodes:                  nodes.Nodes{},
		nodeConnMap:            NodeConnMap{},
		nodeId:                 nodes.NewNodeId(),
	}

	return mgr
}

func (m *MgrNew) Start() {
	go m.eventLoop()
}

func (m *MgrNew) eventLoop() {
	for {
		select {
		case r := <-m.InUiConnectTo:
			m.handleConnectToReq(r)
		case r := <-m.InConnsConnectedStatus:
			m.handleInConnsConnectedStatus(r)
		case r := <-m.IAmReq:
			m.addNodeToCluster(r)
		case r := <-m.myNodesReq:
			m.handleMyNodes(r)
		case r := <-m.saveToClusterReq:
			m.handleSaveToCluster(r)
		case r := <-m.saveToDiskReq:
			m.handleSaveToDisk(r)
		}
	}
}

func (m *MgrNew) handleConnectToReq(i InUiConnectTo) {
	m.OutConnsConnectTo <- OutConnsConnectTo{Address: i.Address}
}

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

type OutConnsConnectTo struct {
	Address string
}

type InConnsConnectedStatus struct {
	Type ConnectedStatus
	Msg  string
	Id   ConnNewId
}
type ConnectedStatus int

const (
	Connected ConnectedStatus = iota
	NotConnected
)

type OutConnsSend struct{}

type InConnsSendStatus struct{}

type InConnsReceive struct{}

type OutDiskSave struct{}

type OutDiskSaveStatus struct{}

func (m *MgrNew) addNodeToCluster(r IAmReq) {
	m.nodes.AddOrUpdate(nodes.NodeNew{Id: r.nodeId})
	m.nodeConnMap.Add(r.nodeId, r.connId)
}
func (m *MgrNew) handleMyNodes(_ MyNodesReq)             {}
func (m *MgrNew) handleSaveToCluster(_ SaveToClusterReq) {}
func (m *MgrNew) handleSaveToDisk(_ SaveToDiskReq)       {}

func (m *MgrNew) handleInConnsConnectedStatus(cs InConnsConnectedStatus) {
	if cs.Type == Connected {
		// Send to OutConnsSend
	}
}
