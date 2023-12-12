package tnet

import (
	"net"
	"tealfs/pkg/model/node"
	"tealfs/pkg/proto"
	"tealfs/pkg/util"
	"time"
)

type Conns struct {
	conns          map[node.Id]Conn
	adds           chan Conn
	deletes        chan node.Id
	tnet           TNet
	myNodeId       node.Id
	connectedNodes chan node.Id
	incoming       chan struct {
		From    node.Id
		Payload proto.Payload
	}
	outgoing chan struct {
		To      node.Id
		Payload proto.Payload
	}
	getlist chan struct {
		response chan util.Set[node.Node]
	}
}

type Conn struct {
	id      node.Id
	address node.Address
	netConn net.Conn
}

func NewConn(address node.Address) Conn {
	return Conn{address: address}
}

func NewConns(tnet TNet, myNodeId node.Id) *Conns {
	conns := &Conns{
		tnet:           tnet,
		myNodeId:       myNodeId,
		conns:          make(map[node.Id]Conn),
		adds:           make(chan Conn),
		deletes:        make(chan node.Id),
		connectedNodes: make(chan node.Id),
		incoming: make(chan struct {
			From    node.Id
			Payload proto.Payload
		}),
		outgoing: make(chan struct {
			To      node.Id
			Payload proto.Payload
		}),
		getlist: make(chan struct {
			response chan util.Set[node.Node]
		}),
	}

	go conns.consumeChannels()
	go conns.acceptConnections()

	return conns
}

func (c *Conns) Add(conn Conn) {
	for {
		if conn.netConn == nil {
			netConn := c.tnet.Dial(conn.address.Value)
			conn = Conn{address: conn.address, netConn: netConn}
		}

		c.sendHello(conn.netConn)
		payload := receivePayload(conn.netConn)
		switch p := payload.(type) {
		case *proto.Hello:
			conn := Conn{address: conn.address, netConn: conn.netConn, id: p.NodeId}
			c.adds <- conn
			return
		}
		conn.netConn.Close()
		conn.netConn = nil

		time.Sleep(time.Second)
	}
}

func (holder *Conns) DeleteConnection(id node.Id) {
	holder.deletes <- id
}

func (holder *Conns) ReceivePayload() (node.Id, proto.Payload) {
	received := <-holder.incoming
	return received.From, received.Payload
}

func (holder *Conns) AddedNode() node.Id {
	return <-holder.connectedNodes
}

func (holder *Conns) SendPayload(to node.Id, payload proto.Payload) {
	holder.outgoing <- struct {
		To      node.Id
		Payload proto.Payload
	}{
		To:      to,
		Payload: payload,
	}
}

func (c *Conns) GetIds() util.Set[node.Id] {
	result := util.NewSet[node.Id]()
	nodes := c.GetNodes()
	for _, node := range nodes.GetValues() {
		result.Add(node.Id)
	}
	return result
}

func (c *Conns) GetNodes() util.Set[node.Node] {
	response := make(chan util.Set[node.Node])
	c.getlist <- struct{ response chan util.Set[node.Node] }{response: response}
	return <-response
}

func (holder *Conns) consumeChannels() {
	for {
		select {
		case conn := <-holder.adds:
			holder.storeNode(conn)

		case id := <-holder.deletes:
			conn, found := holder.conns[id]
			if found {
				if conn.netConn != nil {
					conn.netConn.Close()
				}
				delete(holder.conns, id)
			}

		case sending := <-holder.outgoing:
			netconn := holder.conns[sending.To].netConn
			payload := sending.Payload
			SendPayload(netconn, payload.ToBytes())

		case getList := <-holder.getlist:
			result := util.NewSet[node.Node]()
			for id := range holder.conns {
				conn := holder.conns[id]
				result.Add(node.Node{Id: conn.id, Address: conn.address})
			}
			getList.response <- result
		}
	}
}

func (holder *Conns) storeNode(conn Conn) {
	holder.conns[conn.id] = conn
	holder.connectedNodes <- conn.id
	go holder.readPayloadsFromConnection(conn.id)
}

func (holder *Conns) readPayloadsFromConnection(nodeId node.Id) {
	netConn := holder.netconnForId(nodeId)

	for {
		buf, _ := ReadPayload(netConn)
		payload := proto.ToPayload(buf)
		holder.incoming <- struct {
			From    node.Id
			Payload proto.Payload
		}{From: nodeId, Payload: payload}
	}
}

func (conns *Conns) netconnForId(id node.Id) net.Conn {
	conn := conns.conns[id]
	if conn.netConn != nil {
		return conns.conns[id].netConn
	}
	conn.netConn = conns.tnet.Dial(conn.address.Value)
	return conn.netConn
}

func (c *Conns) acceptConnections() {
	for {
		go c.handleConnection(c.tnet.Accept())
	}
}

func (c *Conns) handleConnection(netConn net.Conn) {
	address := node.NewAddress(netConn.RemoteAddr().String())
	conn := Conn{address: address, netConn: netConn}
	c.Add(conn)
}

func (c *Conns) sendHello(conn net.Conn) {
	hello := proto.Hello{NodeId: c.myNodeId}
	SendPayload(conn, hello.ToBytes())
}

func receivePayload(conn net.Conn) proto.Payload {
	bytes, _ := ReadPayload(conn)
	return proto.ToPayload(bytes)
}
