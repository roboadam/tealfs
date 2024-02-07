package tnet

import (
	"net"
	"tealfs/pkg/model/node"
	"tealfs/pkg/proto"
	"tealfs/pkg/set"
	"time"
)

type Conns struct {
	conns          map[node.Id]Conn
	adds           chan Conn
	deletes        chan node.Id
	tnet           TNet
	MyNodeId       node.Id
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
		response chan set.Set[node.Node]
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
		MyNodeId:       myNodeId,
		conns:          make(map[node.Id]Conn),
		adds:           make(chan Conn, 100),
		deletes:        make(chan node.Id, 100),
		connectedNodes: make(chan node.Id, 100),
		incoming: make(chan struct {
			From    node.Id
			Payload proto.Payload
		}, 100),
		outgoing: make(chan struct {
			To      node.Id
			Payload proto.Payload
		}, 100),
		getlist: make(chan struct {
			response chan set.Set[node.Node]
		}, 100),
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
		_ = conn.netConn.Close()
		conn.netConn = nil

		time.Sleep(time.Second)
	}
}

func (c *Conns) DeleteConnection(id node.Id) {
	c.deletes <- id
}

func (c *Conns) ReceivePayload() (node.Id, proto.Payload) {
	received := <-c.incoming
	return received.From, received.Payload
}

func (c *Conns) AddedNode() node.Id {
	return <-c.connectedNodes
}

func (c *Conns) SendPayload(to node.Id, payload proto.Payload) {
	c.outgoing <- struct {
		To      node.Id
		Payload proto.Payload
	}{
		To:      to,
		Payload: payload,
	}
}

func (c *Conns) GetIds() set.Set[node.Id] {
	result := set.NewSet[node.Id]()
	nodes := c.GetNodes()
	for _, n := range nodes.GetValues() {
		result.Add(n.Id)
	}
	return result
}

func (c *Conns) GetNodes() set.Set[node.Node] {
	response := make(chan set.Set[node.Node])
	c.getlist <- struct{ response chan set.Set[node.Node] }{response: response}
	result := <-response
	return result
}

func (c *Conns) consumeChannels() {
	for {
		select {
		case conn := <-c.adds:
			c.storeNode(conn)

		case id := <-c.deletes:
			conn, found := c.conns[id]
			if found {
				if conn.netConn != nil {
					_ = conn.netConn.Close()
				}
				delete(c.conns, id)
			}

		case sending := <-c.outgoing:
			netconn := c.conns[sending.To].netConn
			payload := sending.Payload
			_ = SendPayload(netconn, payload.ToBytes())

		case getList := <-c.getlist:
			result := set.NewSet[node.Node]()
			for id := range c.conns {
				conn := c.conns[id]
				result.Add(node.Node{Id: conn.id, Address: conn.address})
			}
			getList.response <- result
		default:
			//do nothing
		}
	}
}

func (c *Conns) storeNode(conn Conn) {
	c.conns[conn.id] = conn
	c.connectedNodes <- conn.id
	go c.readPayloadsFromConnection(conn.id)
}

func (c *Conns) readPayloadsFromConnection(nodeId node.Id) {
	netConn := c.netconnForId(nodeId)

	for {
		payload := receivePayload(netConn)
		c.incoming <- struct {
			From    node.Id
			Payload proto.Payload
		}{From: nodeId, Payload: payload}
	}
}

func (c *Conns) netconnForId(id node.Id) net.Conn {
	conn := c.conns[id]
	if conn.netConn != nil {
		return c.conns[id].netConn
	}
	conn.netConn = c.tnet.Dial(conn.address.Value)
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
	hello := proto.Hello{NodeId: c.MyNodeId}
	_ = SendPayload(conn, hello.ToBytes())
}

func receivePayload(conn net.Conn) proto.Payload {
	bytes, _ := ReadPayload(conn)
	return proto.ToPayload(bytes)
}
