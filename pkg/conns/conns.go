package conns

import (
	"net"
	"tealfs/pkg/node"
	"tealfs/pkg/proto"
	"tealfs/pkg/raw_net"
	"tealfs/pkg/tnet"
	"time"
)

type Conns struct {
	conns    map[node.Id]Conn
	adds     chan Conn
	deletes  chan node.Id
	tnet     tnet.TNet
	myNodeId node.Id
	incoming chan struct {
		From    node.Id
		Payload proto.Payload
	}
	outgoing chan struct {
		To      node.Id
		Payload proto.Payload
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

func New(tnet tnet.TNet, myNodeId node.Id) *Conns {
	conns := &Conns{
		tnet:     tnet,
		myNodeId: myNodeId,
		conns:    make(map[node.Id]Conn),
		adds:     make(chan Conn),
		deletes:  make(chan node.Id),
		incoming: make(chan struct {
			From    node.Id
			Payload proto.Payload
		}),
		outgoing: make(chan struct {
			To      node.Id
			Payload proto.Payload
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

func (holder *Conns) SendPayload(to node.Id, payload proto.Payload) {
	holder.outgoing <- struct {
		To      node.Id
		Payload proto.Payload
	}{
		To:      to,
		Payload: payload,
	}
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
			raw_net.SendPayload(netconn, payload.ToBytes())
		}
	}
}

func (holder *Conns) storeNode(conn Conn) {
	holder.conns[conn.id] = conn
	go holder.readPayloadsFromConnection(conn.id)
}

func (holder *Conns) readPayloadsFromConnection(nodeId node.Id) {
	netConn := holder.netconnForId(nodeId)

	for {
		buf, _ := raw_net.ReadPayload(netConn)
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
	raw_net.SendPayload(conn, hello.ToBytes())
}

func receivePayload(conn net.Conn) proto.Payload {
	bytes, _ := raw_net.ReadPayload(conn)
	return proto.ToPayload(bytes)
}
