package mgr

import (
	"tealfs/pkg/nodes"
	"tealfs/pkg/proto"
)

type MgrNew struct {
	ConnToReq        chan ConnectToReq
	incomingConnReq  <-chan IncomingConnReq
	iAmReq           <-chan IAmReq
	myNodesReq       <-chan MyNodesReq
	saveToClusterReq <-chan SaveToClusterReq
	saveToDiskReq    <-chan SaveToDiskReq

	conns       ConnsNew
	nodes       nodes.Nodes
	nodeConnMap NodeConnMap
	nodeId      nodes.Id
}

func NewNew() MgrNew {
	iAmReq := make(chan IAmReq, 100)
	incomingConnReq := make(chan IncomingConnReq, 100)
	connToReq := make(chan ConnectToReq, 100)
	mgr := MgrNew{
		ConnToReq:        connToReq,
		incomingConnReq:  incomingConnReq,
		iAmReq:           iAmReq,
		myNodesReq:       make(chan MyNodesReq, 100),
		saveToClusterReq: make(chan SaveToClusterReq, 100),
		saveToDiskReq:    make(chan SaveToDiskReq, 100),
		conns:            ConnsNewNew(iAmReq, incomingConnReq),
		nodeConnMap:      NodeConnMap{},
		nodeId:           nodes.NewNodeId(),
	}

	return mgr
}

func (m *MgrNew) Start() {
	go m.eventLoop()
}

func (m *MgrNew) eventLoop() {
	for {
		select {
		case r := <-m.ConnToReq:
			m.handleConnectToReq(r)
		case r := <-m.incomingConnReq:
			m.conns.SaveIncoming(r)
		case r := <-m.iAmReq:
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

func (m *MgrNew) handleConnectToReq(r ConnectToReq) {
	id, err := m.conns.ConnectTo(r.Address)
	if err == nil {
		iam := &proto.IAm{NodeId: m.nodeId}
		err = m.conns.Send(id, iam.ToBytes())
	}
	r.Resp <- ConnectToResp{Id: id, Success: err == nil}
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

func (m *MgrNew) addNodeToCluster(r IAmReq) {
	m.nodes.AddOrUpdate(nodes.NodeNew{Id: r.nodeId})
	m.nodeConnMap.Add(r.nodeId, r.connId)
}
func (m *MgrNew) handleMyNodes(_ MyNodesReq)             {}
func (m *MgrNew) handleSaveToCluster(_ SaveToClusterReq) {}
func (m *MgrNew) handleSaveToDisk(_ SaveToDiskReq)       {}
