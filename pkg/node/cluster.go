package node

import (
	"fmt"
	"net"
)

type NodeConnection struct {
	Address string
	Conn    net.Conn
}

type NodeConnectionId struct {
	Value uint32
}

type Connections struct {
	connections map[uint32]NodeConnection
	addChan     chan struct {
		id   uint32
		conn NodeConnection
	}
	getChan chan struct {
		id       uint32
		response chan *NodeConnection
	}
	deleteChan chan uint32
	nextIdx    uint32
}

func NewConnections() *Connections {
	connections := &Connections{
		connections: make(map[uint32]NodeConnection),
		addChan: make(chan struct {
			id   uint32
			conn NodeConnection
		}),
		getChan: make(chan struct {
			id       uint32
			response chan *NodeConnection
		}),
		deleteChan: make(chan uint32),
		nextIdx:    0,
	}

	go connections.run()

	return connections
}

func (holder *Connections) run() {
	for {
		select {
		case request := <-holder.addChan:
			holder.connections[request.id] = request.conn

		case request := <-holder.getChan:
			conn, found := holder.connections[request.id]
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

func (holder *Connections) AddConnection(conn NodeConnection) NodeConnectionId {
	id := holder.nextIdx
	holder.nextIdx++
	holder.addChan <- struct {
		id   uint32
		conn NodeConnection
	}{id, conn}

	return NodeConnectionId{Value: id}
}

func (holder *Connections) GetConnection(id NodeConnectionId) *NodeConnection {
	responseChan := make(chan *NodeConnection)
	holder.getChan <- struct {
		id       uint32
		response chan *NodeConnection
	}{id.Value, responseChan}
	return <-responseChan
}

func (holder *Connections) DeleteConnection(id NodeConnectionId) {
	holder.deleteChan <- id.Value
}

func (holder *Connections) ConnectAll() {
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
