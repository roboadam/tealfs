package node

type RemoteNodes struct {
	nodes   map[NodeId]RemoteNode
	adds    chan RemoteNode
	gets    chan getsRequestWithResponseChan
	deletes chan NodeId
}

func NewRemoteNodes() *RemoteNodes {
	nodes := &RemoteNodes{
		nodes:   make(map[NodeId]RemoteNode),
		adds:    make(chan RemoteNode),
		gets:    make(chan getsRequestWithResponseChan),
		deletes: make(chan NodeId),
	}

	go nodes.run()

	return nodes
}

type getsRequestWithResponseChan struct {
	request  NodeId
	response chan *RemoteNode
}

func (holder *RemoteNodes) run() {
	for {
		select {
		case request := <-holder.adds:
			holder.storeNode(request)

		case request := <-holder.gets:
			holder.sendConnectionToChan(request)

		case id := <-holder.deletes:
			delete(holder.nodes, id)
		}
	}
}

func (holder *RemoteNodes) sendConnectionToChan(request getsRequestWithResponseChan) {
	conn, found := holder.nodes[request.request]
	if found {
		request.response <- &conn
	} else {
		request.response <- nil
	}
}

func (holder *RemoteNodes) storeNode(request RemoteNode) {
	request.Connect()
	holder.nodes[request.NodeId] = request
}

func (holder *RemoteNodes) AddConnection(node RemoteNode) {
	holder.adds <- node
}

func (holder *RemoteNodes) GetConnection(id NodeId) *RemoteNode {
	responseChan := make(chan *RemoteNode)
	holder.gets <- getsRequestWithResponseChan{id, responseChan}
	return <-responseChan
}

func (holder *RemoteNodes) DeleteConnection(id NodeId) {
	holder.deletes <- id
}

func (holder *RemoteNodes) ConnectAll() {
	for _, node := range holder.nodes {
		node.Connect()
	}
}
