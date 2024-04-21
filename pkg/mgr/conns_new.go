package mgr

import (
	"net"
	"tealfs/pkg/proto"
	"tealfs/pkg/tnet"
)

type ConnId int32
type Conns struct {
	netConns               map[ConnId]net.Conn
	nextId                 ConnId
	outConnsConnectTo      <-chan MgrConnsConnectTo
	inConnsConnectedStatus chan<- ConnsMgrStatus
	incomingConnReq        chan<- IncomingConnReq
	listener               net.Listener
}

func ConnsNew(outConnsConnectTo <-chan MgrConnsConnectTo, status chan<- ConnsMgrStatus) Conns {
	listener, err := net.Listen("tcp", "localhost:0")
	if err != nil {
		panic(err)
	}
	c := Conns{
		netConns:               make(map[ConnId]net.Conn, 3),
		nextId:                 ConnId(0),
		outConnsConnectTo:      outConnsConnectTo,
		inConnsConnectedStatus: status,
	}
	go c.listen(listener)
	return c
}

func (c *Conns) consumeChannels() {
	select {
	case connectToReq := <-c.outConnsConnectTo:
		id, err := c.connectTo(connectToReq.Address)
		if err == nil {
			c.inConnsConnectedStatus <- ConnsMgrStatus{
				Type: Connected,
				Msg:  "Success",
				Id:   id,
			}
		}

	}
}

func (c *Conns) listen(listener net.Listener) {
	for {
		conn, err := listener.Accept()
		if err == nil {
			incomingConnReq := IncomingConnReq{netConn: conn}
			c.incomingConnReq <- incomingConnReq
		}
	}
}

type UiMgrConnectTo struct {
	Address string
}
type ConnectToResp struct {
	Success      bool
	Id           ConnId
	ErrorMessage string
}

type IncomingConnReq struct {
	netConn net.Conn
}

func (c *Conns) consumeData(conn ConnId) {
	for {
		netConn := c.netConns[conn]
		bytes, err := tnet.ReadPayload(netConn)
		if err != nil {
			return
		}
		payload := proto.ToPayload(bytes)
		switch _ := payload.(type) {
		case *proto.SyncNodes:
			break
		case *proto.SaveData:
			break
		default:
			// Do nothing
		}
	}
}

func (c *Conns) SaveIncoming(req IncomingConnReq) {
	_ = c.saveNetConn(req.netConn)
}

func (c *Conns) connectTo(address string) (ConnId, error) {
	netConn, err := net.Dial("tcp", address)
	if err != nil {
		return 0, err
	}
	id := c.saveNetConn(netConn)
	return id, nil
}

func (c *Conns) Send(id ConnId, data []byte) error {
	bytesWritten := 0
	for bytesWritten < len(data) {
		n, err := c.netConns[id].Write(data[bytesWritten:])
		if err != nil {
			return err
		}
		bytesWritten += n
	}
	return nil
}

func (c *Conns) saveNetConn(netConn net.Conn) ConnId {
	id := c.nextId
	c.nextId++
	c.netConns[id] = netConn
	return id
}
