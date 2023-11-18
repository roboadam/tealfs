package node

import (
	"errors"
	"net"
	"tealfs/pkg/raw_net"
	"tealfs/pkg/util"
)

type RemoteNodes struct {
	nodes     map[Id]remoteNode
	localNode Node
	adds      chan remoteNode
	gets      chan getsRequestWithResponseChan
	deletes   chan Id
	incoming  chan struct {
		From    Id
		Payload *Payload
	}
	outgoing chan struct {
		To      Id
		Payload *Payload
	}
}

func NewRemoteNodes(localNode Node) *RemoteNodes {
	nodes := &RemoteNodes{
		nodes:     make(map[Id]remoteNode),
		localNode: localNode,
		adds:      make(chan remoteNode),
		gets:      make(chan getsRequestWithResponseChan),
		deletes:   make(chan Id),
		incoming: make(chan struct {
			From    Id
			Payload *Payload
		}),
		outgoing: make(chan struct {
			To      Id
			Payload *Payload
		}),
	}

	go nodes.consumeChannels()

	return nodes
}

func (holder *RemoteNodes) Add(node Node, conn net.Conn) {
	holder.adds <- remoteNode{node: node, conn: conn}
}

func (holder *RemoteNodes) GetNode(id Id) (*Node, error) {
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

func (holder *RemoteNodes) ReceivePayload() (Id, *Payload) {
	received := <-holder.incoming
	return received.From, received.Payload
}

func (holder *RemoteNodes) SendPayload(to Id, payload *Payload) {
	holder.outgoing <- struct {
		To      Id
		Payload *Payload
	}{
		To:      to,
		Payload: payload,
	}
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

		case sending := <-holder.outgoing:
			conn := holder.nodes[sending.To].conn
			payload := *sending.Payload
			raw_net.SendPayload(conn, payload.ToBytes())
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
	go holder.readPayloadsFromConnection(remoteNode.node.Id)
}

func (holder *RemoteNodes) nodesToSync() util.Set[Node] {
	result := util.NewSet[Node]()

	result.Add(holder.localNode)
	for _, remoteNode := range holder.nodes {
		result.Add(remoteNode.node)
	}

	return result
}

func (holder *RemoteNodes) readPayloadsFromConnection(nodeId Id) {
	conn := holder.nodes[nodeId].conn
	for {
		buf, _ := raw_net.ReadPayload(conn)
		payload := ToPayload(buf)
		holder.incoming <- struct {
			From    Id
			Payload *Payload
		}{From: nodeId, Payload: &payload}
	}
}

type remoteNode struct {
	node Node
	conn net.Conn
}
