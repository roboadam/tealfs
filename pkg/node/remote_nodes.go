package node

import "errors"

type RemoteNodes struct {
	nodes   map[Id]RemoteNode
	adds    chan RemoteNode
	gets    chan getsRequestWithResponseChan
	deletes chan Id
}

func NewRemoteNodes() *RemoteNodes {
	nodes := &RemoteNodes{
		nodes:   make(map[Id]RemoteNode),
		adds:    make(chan RemoteNode),
		gets:    make(chan getsRequestWithResponseChan),
		deletes: make(chan Id),
	}

	go nodes.consumeChannels()

	return nodes
}

func (holder *RemoteNodes) Add(node RemoteNode) {
	holder.adds <- node
}

func (holder *RemoteNodes) GetConnection(id Id) (*RemoteNode, error) {
	responseChan := make(chan *RemoteNode)
	holder.gets <- getsRequestWithResponseChan{id, responseChan}
	conn := <-responseChan
	if conn == nil {
		return conn, errors.New("No connection with ID " + id.String())
	}
	return conn, nil
}

func (holder *RemoteNodes) DeleteConnection(id Id) {
	holder.deletes <- id
}

type getsRequestWithResponseChan struct {
	request  Id
	response chan *RemoteNode
}

func (holder *RemoteNodes) consumeChannels() {
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

func (holder *RemoteNodes) storeNode(node RemoteNode) {
	holder.nodes[node.Id] = node
}
