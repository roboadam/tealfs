package mgr

type MgrNew struct {
	connToReq        <-chan ConnectToReq
	incomingConnReq  <-chan IncomingConnReq
	iAmReq           <-chan IAmReq
	myNodesReq       <-chan MyNodesReq
	saveToClusterReq <-chan SaveToClusterReq
	saveToDiskReq    <-chan SaveToDiskReq
}

func (m *MgrNew) Start() {
	for {
		select {
		case r := <-m.connToReq:
			m.handleConnTo(r)
		case r := <-m.incomingConnReq:
			m.handleIncomingConn(r)
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

type ConnectToReq struct {
	resp chan<- ConnectToResp
}
type ConnectToResp struct {
}

type IncomingConnReq struct {
	resp chan<- IncomingConnResp
}
type IncomingConnResp struct {
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

func (m *MgrNew) handleConnTo(_ ConnectToReq)            {}
func (m *MgrNew) handleIncomingConn(_ IncomingConnReq)   {}
func (m *MgrNew) handleIAm(_ IAmReq)                     {}
func (m *MgrNew) handleMyNodes(_ MyNodesReq)             {}
func (m *MgrNew) handleSaveToCluster(_ SaveToClusterReq) {}
func (m *MgrNew) handleSaveToDisk(_ SaveToDiskReq)       {}
