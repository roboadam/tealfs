package mgr

type MgrNew struct {
	connToReq        <-chan ConnectToReq
	incomingConnReq  <-chan IncomingConnReq
	iAmReq           <-chan IAmReq
	myNodesReq       <-chan MyNodesReq
	saveToClusterReq <-chan SaveToClusterReq
	saveToDiskReq    <-chan SaveToDiskReq

	conns ConnsNew
	nodes NodesNew
}

func (m *MgrNew) Start() {
	go m.eventLoop()
}

func (m *MgrNew) eventLoop() {
	for {
		select {
		case r := <-m.connToReq:
			m.conns.ConnectTo(r)
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

type IAmReq struct {
	resp chan<- IAmResp
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
