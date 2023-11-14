package node

import (
	"errors"
	"net"
	"tealfs/pkg/proto"
	"tealfs/pkg/raw_net"
)

type RemoteNodes struct {
	nodes    map[Id]remoteNode
	adds     chan remoteNode
	gets     chan getsRequestWithResponseChan
	deletes  chan Id
	incoming chan struct {
		From Id
		Data []byte
	}
	outgoing chan struct {
		To   Id
		Data []byte
	}
}

func NewRemoteNodes() *RemoteNodes {
	nodes := &RemoteNodes{
		nodes:   make(map[Id]remoteNode),
		adds:    make(chan remoteNode),
		gets:    make(chan getsRequestWithResponseChan),
		deletes: make(chan Id),
		incoming: make(chan struct {
			From Id
			Data []byte
		}),
		outgoing: make(chan struct {
			To   Id
			Data []byte
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

func (holder *RemoteNodes) ReceivePayload() (Id, *proto.Payload) {
	received := <-holder.incoming
	return received.From, received.Data
}

func (holder *RemoteNodes) SendPayload(to Id, payload *proto.Payload) {
	holder.outgoing <- struct {
		To      Id
		Payload *proto.Payload
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
			raw_net.SendBytes(conn, sending.Data.Data)
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

func (holder *RemoteNodes) readPayloadsFromConnection(nodeId Id) {
	conn := holder.nodes[nodeId].conn
	for {
		buf, _ := raw_net.ReadBytes(conn, proto.CommandAndLengthSize)
		cmd, len, _ := proto.CommandAndLengthFromBytes(buf)
		data, _ := raw_net.ReadBytes(conn, len)
		holder.incoming <- &Payload{NodeId: nodeId, Command: cmd, RawData: data}
	}
}

type remoteNode struct {
	node Node
	conn net.Conn
}
