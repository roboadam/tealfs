package node

import (
	"fmt"
	"net"
)

type RemoteNodes struct {
	connections map[NodeId]RemoteNode
	addChan     chan RemoteNode
	getChan     chan struct {
		request  NodeId
		response chan *RemoteNode
	}
	deleteChan chan NodeId
}

func NewRemoteNodes() *RemoteNodes {
	nodes := &RemoteNodes{
		connections: make(map[NodeId]RemoteNode),
		addChan:     make(chan RemoteNode),
		getChan: make(chan struct {
			request  NodeId
			response chan *RemoteNode
		}),
		deleteChan: make(chan NodeId),
	}

	go nodes.run()

	return nodes
}

func (holder *RemoteNodes) run() {
	for {
		select {
		case request := <-holder.addChan:
			holder.connections[request.NodeId] = request

		case request := <-holder.getChan:
			conn, found := holder.connections[request.request]
			if found {
				request.response <- &conn
			} else {
				request.response <- nil
			}

		case id := <-holder.deleteChan:
			delete(holder.connections, id)
		}
	}
}

func (holder *RemoteNodes) AddConnection(node RemoteNode) {
	holder.addChan <- node
}

func (holder *RemoteNodes) GetConnection(id NodeId) *RemoteNode {
	responseChan := make(chan *RemoteNode)
	holder.getChan <- struct {
		request  NodeId
		response chan *RemoteNode
	}{id, responseChan}
	return <-responseChan
}

func (holder *RemoteNodes) DeleteConnection(id NodeId) {
	holder.deleteChan <- id
}

func (holder *RemoteNodes) ConnectAll() {
	for _, conn := range holder.connections {
		if conn.Conn == nil {
			tcpConn, error := net.Dial("tcp", conn.Address)
			if error == nil {
				fmt.Println("Connected to somebody")
				conn.Conn = tcpConn
			} else {
				fmt.Println("Cant connect!")
			}
		}
	}
}
