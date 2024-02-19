package tnet

import (
	"net"
	"tealfs/pkg/nodes"
	"tealfs/pkg/proto"
	"tealfs/pkg/set"
	"time"
)

type Conns struct {
	conns          map[nodes.Id]Conn
	adds           chan Conn
	deletes        chan nodes.Id
	tnet           TNet
	MyNodeId       nodes.Id
	connectedNodes chan nodes.Id
	incoming       chan struct {
		From    nodes.Id
		Payload proto.Payload
	}
	outgoing chan struct {
		To      nodes.Id
		Payload proto.Payload
	}
	getlist chan struct {
		response chan set.Set[nodes.Node]
	}
}

type Conn struct {
	id      nodes.Id
	address nodes.Address
	netConn net.Conn
}

func NewConn(address nodes.Address) Conn {
	return Conn{address: address}
}

func NewConns(tnet TNet, myNodeId nodes.Id) *Conns {
	conns := &Conns{
		tnet:           tnet,
		MyNodeId:       myNodeId,
		conns:          make(map[nodes.Id]Conn),
		adds:           make(chan Conn, 100),
		deletes:        make(chan nodes.Id, 100),
		connectedNodes: make(chan nodes.Id, 100),
		incoming: make(chan struct {
			From    nodes.Id
			Payload proto.Payload
		}, 100),
		outgoing: make(chan struct {
			To      nodes.Id
			Payload proto.Payload
		}, 100),
		getlist: make(chan struct {
			response chan set.Set[nodes.Node]
		}, 100),
	}

	go conns.consumeChannels()
	go conns.acceptConnections()

	return conns
}

func (c *Conns) Add(conn Conn) {
	for {
		if conn.netConn == nil {
			netConn := c.tnet.Dial(string(conn.address))
			conn = Conn{address: conn.address, netConn: netConn}
		}

		c.sendHello(conn.netConn)
		payload := receivePayload(conn.netConn)
		switch p := payload.(type) {
		case *proto.IAm:
			conn := Conn{address: conn.address, netConn: conn.netConn, id: p.NodeId}
			c.adds <- conn
			return
		}
		_ = conn.netConn.Close()
		conn.netConn = nil

		time.Sleep(time.Second)
	}
}

func (c *Conns) DeleteConnection(id nodes.Id) {
	c.deletes <- id
}

func (c *Conns) ReceivePayload() (nodes.Id, proto.Payload) {
	received := <-c.incoming
	return received.From, received.Payload
}

func (c *Conns) AddedNode() nodes.Id {
	return <-c.connectedNodes
}

func (c *Conns) SendPayload(to nodes.Id, payload proto.Payload) {
	c.outgoing <- struct {
		To      nodes.Id
		Payload proto.Payload
	}{
		To:      to,
		Payload: payload,
	}
}

func (c *Conns) GetIds() set.Set[nodes.Id] {
	result := set.NewSet[nodes.Id]()
	nds := c.GetNodes()
	for _, n := range nds.GetValues() {
		result.Add(n.Id)
	}
	return result
}

func (c *Conns) GetNodes() set.Set[nodes.Node] {
	response := make(chan set.Set[nodes.Node])
	c.getlist <- struct{ response chan set.Set[nodes.Node] }{response: response}
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
			result := set.NewSet[nodes.Node]()
			for id := range c.conns {
				conn := c.conns[id]
				result.Add(nodes.Node{Id: conn.id, Address: conn.address})
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

func (c *Conns) readPayloadsFromConnection(nodeId nodes.Id) {
	netConn := c.netconnForId(nodeId)

	for {
		payload := receivePayload(netConn)
		c.incoming <- struct {
			From    nodes.Id
			Payload proto.Payload
		}{From: nodeId, Payload: payload}
	}
}

func (c *Conns) netconnForId(id nodes.Id) net.Conn {
	conn := c.conns[id]
	if conn.netConn != nil {
		return c.conns[id].netConn
	}
	conn.netConn = c.tnet.Dial(string(conn.address))
	return conn.netConn
}

func (c *Conns) acceptConnections() {
	for {
		go c.handleConnection(c.tnet.Accept())
	}
}

func (c *Conns) handleConnection(netConn net.Conn) {
	address := nodes.Address(netConn.RemoteAddr().String())
	conn := Conn{address: address, netConn: netConn}
	c.Add(conn)
}

func (c *Conns) sendHello(conn net.Conn) {
	hello := proto.IAm{NodeId: c.MyNodeId}
	_ = SendPayload(conn, hello.ToBytes())
}

func receivePayload(conn net.Conn) proto.Payload {
	bytes, _ := ReadPayload(conn)
	return proto.ToPayload(bytes)
}
