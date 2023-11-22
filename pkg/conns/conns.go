package conns

import (
	"net"
	"tealfs/pkg/node"
	"tealfs/pkg/raw_net"
	"tealfs/pkg/tnet"
)

type Conns struct {
	conns    map[node.Id]conn
	adds     chan conn
	deletes  chan node.Id
	tnet     tnet.TNet
	incoming chan struct {
		From    node.Id
		Payload *node.Payload
	}
	outgoing chan struct {
		To      node.Id
		Payload *node.Payload
	}
}

type conn struct {
	id      node.Id
	address node.Address
	netConn net.Conn
}

func New(tnet tnet.TNet) *Conns {
	conns := &Conns{
		tnet:    tnet,
		conns:   make(map[node.Id]conn),
		adds:    make(chan conn),
		deletes: make(chan node.Id),
		incoming: make(chan struct {
			From    node.Id
			Payload *node.Payload
		}),
		outgoing: make(chan struct {
			To      node.Id
			Payload *node.Payload
		}),
	}

	go conns.consumeChannels()

	return conns
}

func (holder *Conns) Add(id node.Id, address node.Address) {
	holder.adds <- conn{id: id, address: address}
}

func (holder *Conns) DeleteConnection(id node.Id) {
	holder.deletes <- id
}

func (holder *Conns) ReceivePayload() (node.Id, *node.Payload) {
	received := <-holder.incoming
	return received.From, received.Payload
}

func (holder *Conns) SendPayload(to node.Id, payload *node.Payload) {
	holder.outgoing <- struct {
		To      node.Id
		Payload *node.Payload
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
			payload := *sending.Payload
			raw_net.SendPayload(netconn, payload.ToBytes())
		}
	}
}

func (holder *Conns) storeNode(conn conn) {
	holder.conns[conn.id] = conn
	go holder.readPayloadsFromConnection(conn.id)
}

func (holder *Conns) readPayloadsFromConnection(nodeId node.Id) {
	netConn := holder.netconnForId(nodeId)

	for {
		buf, _ := raw_net.ReadPayload(netConn)
		payload := node.ToPayload(buf)
		holder.incoming <- struct {
			From    node.Id
			Payload *node.Payload
		}{From: nodeId, Payload: &payload}
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
