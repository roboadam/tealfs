package node

import (
	"errors"
	"net"
)

type RemoteNodes struct {
	nodes   map[Id]remoteNode
	adds    chan remoteNode
	gets    chan getsRequestWithResponseChan
	deletes chan Id
}

type remoteNode struct {
	node Node
	conn net.Conn
}

func NewRemoteNodes() *RemoteNodes {
	nodes := &RemoteNodes{
		nodes:   make(map[Id]remoteNode),
		adds:    make(chan remoteNode),
		gets:    make(chan getsRequestWithResponseChan),
		deletes: make(chan Id),
	}

	go nodes.consumeChannels()

	return nodes
}

func (holder *RemoteNodes) Add(node Node, conn net.Conn) {
	holder.adds <- remoteNode{node: node, conn: conn}
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
			remoteNode, found := holder.nodes[id]
			if found {
				remoteNode.conn.Close()
				delete(holder.nodes, id)
			}
		}
	}
}

func (holder *RemoteNodes) sendNodeToChan(request getsRequestWithResponseChan) {
	remoteNode, found := holder.nodes[request.request]
	if found {
		request.response <- &remoteNode.node
	} else {
		request.response <- nil
	}
}

func (holder *RemoteNodes) storeNode(remoteNode remoteNode) {
	holder.nodes[remoteNode.node.Id] = remoteNode
}
