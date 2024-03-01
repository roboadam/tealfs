package mgr

import (
	"tealfs/pkg/nodes"
	"tealfs/pkg/proto"
)

type MgrNew struct {
	connToReq        chan ConnectToReq
	incomingConnReq  <-chan IncomingConnReq
	iAmReq           <-chan IAmReq
	myNodesReq       <-chan MyNodesReq
	saveToClusterReq <-chan SaveToClusterReq
	saveToDiskReq    <-chan SaveToDiskReq

	conns  ConnsNew
	nodes  nodes.Nodes
	nodeId nodes.Id
}

func NewNew() MgrNew {
	iAmReq := make(chan IAmReq, 100)
	mgr := MgrNew{
		connToReq:        make(chan ConnectToReq, 100),
		incomingConnReq:  make(chan IncomingConnReq, 100),
		iAmReq:           iAmReq,
		myNodesReq:       make(chan MyNodesReq, 100),
		saveToClusterReq: make(chan SaveToClusterReq, 100),
		saveToDiskReq:    make(chan SaveToDiskReq, 100),
		conns:            ConnsNewNew(iAmReq),
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
		case r := <-m.connToReq:
			m.handleConnectToReq(r)
		case r := <-m.incomingConnReq:
			m.conns.HandleIncoming(r)
		case r := <-m.iAmReq:
			m.handleIAm(r)
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
	id, err := m.conns.ConnectTo(r.address)
	if err == nil {
		iam := &proto.IAm{NodeId: m.nodeId}
		err = m.conns.Send(id, iam.ToBytes())
	}
	r.resp <- ConnectToResp{Id: id, Success: err == nil}
}

type IAmReq struct {
	nodeId nodes.Id
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

func (m *MgrNew) handleIAm(_ IAmReq)                     {}
func (m *MgrNew) handleMyNodes(_ MyNodesReq)             {}
func (m *MgrNew) handleSaveToCluster(_ SaveToClusterReq) {}
func (m *MgrNew) handleSaveToDisk(_ SaveToDiskReq)       {}
