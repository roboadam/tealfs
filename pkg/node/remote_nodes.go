package node

import "errors"

type RemoteNodes struct {
	nodes   map[Id]Node
	adds    chan Node
	gets    chan getsRequestWithResponseChan
	deletes chan Id
}

func NewRemoteNodes() *RemoteNodes {
	nodes := &RemoteNodes{
		nodes:   make(map[Id]Node),
		adds:    make(chan Node),
		gets:    make(chan getsRequestWithResponseChan),
		deletes: make(chan Id),
	}

	go nodes.consumeChannels()

	return nodes
}

func (holder *RemoteNodes) Add(node Node) {
	holder.adds <- node
}

func (holder *RemoteNodes) GetConnection(id Id) (*Node, error) {
	responseChan := make(chan *Node)
	holder.gets <- getsRequestWithResponseChan{id, responseChan}
	node := <-responseChan
	if node == nil {
		return node, errors.New("No connection with ID " + id.String())
	}
	return node, nil
}

func (holder *RemoteNodes) DeleteConnection(id Id) {
	holder.deletes <- id
}

type getsRequestWithResponseChan struct {
	request  Id
	response chan *Node
}

func (holder *RemoteNodes) consumeChannels() {
	for {
		select {
		case request := <-holder.adds:
			holder.storeNode(request)

		case request := <-holder.gets:
			holder.sendNodeToChan(request)

		case id := <-holder.deletes:
			delete(holder.nodes, id)
		}
	}
}

func (holder *RemoteNodes) sendNodeToChan(request getsRequestWithResponseChan) {
	conn, found := holder.nodes[request.request]
	if found {
		request.response <- &conn
	} else {
		request.response <- nil
	}
}

func (holder *RemoteNodes) storeNode(node Node) {
	holder.nodes[node.Id] = node
}
