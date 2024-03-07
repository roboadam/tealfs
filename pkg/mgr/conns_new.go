package mgr

import (
	"net"
	"tealfs/pkg/proto"
	"tealfs/pkg/tnet"
)

type ConnNewId int32
type ConnsNew struct {
	netConns        map[ConnNewId]net.Conn
	nextId          ConnNewId
	iAmReq          chan<- IAmReq
	incomingConnReq chan<- IncomingConnReq
	listener        net.Listener
}

func ConnsNewNew(iamReqs chan<- IAmReq, incomingConnReq chan<- IncomingConnReq) ConnsNew {
	listener, err := net.Listen("tcp", "localhost:0")
	if err != nil {
		panic(err)
	}
	c := ConnsNew{
		netConns:        make(map[ConnNewId]net.Conn, 3),
		nextId:          ConnNewId(0),
		iAmReq:          iamReqs,
		incomingConnReq: incomingConnReq,
	}
	c.listen(listener)
	return c
}

func (c *ConnsNew) listen(listener net.Listener) {
	for {
		conn, err := listener.Accept()
		if err == nil {
			incomingConnReq := IncomingConnReq{netConn: conn}
			c.incomingConnReq <- incomingConnReq
		}
	}
}

type ConnectToReq struct {
	address string
	resp    chan<- ConnectToResp
}
type ConnectToResp struct {
	Success      bool
	Id           ConnNewId
	ErrorMessage string
}

type IncomingConnReq struct {
	netConn net.Conn
}

func (c *ConnsNew) consumeData(conn ConnNewId) {
	for {
		netConn := c.netConns[conn]
		bytes, err := tnet.ReadPayload(netConn)
		if err != nil {
			return
		}
		payload := proto.ToPayload(bytes)
		switch p := payload.(type) {
		case *proto.IAm:
			c.iAmReq <- IAmReq{nodeId: p.NodeId}
		case *proto.SyncNodes:
			break
		case *proto.SaveData:
			break
		default:
			// Do nothing
		}
	}
}

func (c *ConnsNew) SaveIncoming(req IncomingConnReq) {
	_ = c.saveNetConn(req.netConn)
}

func (c *ConnsNew) ConnectTo(address string) (ConnNewId, error) {
	netConn, err := net.Dial("tcp", address)
	if err != nil {
		return 0, err
	}
	id := c.saveNetConn(netConn)
	return id, nil
}

func (c *ConnsNew) Send(id ConnNewId, data []byte) error {
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

func (c *ConnsNew) saveNetConn(netConn net.Conn) ConnNewId {
	id := c.nextId
	c.nextId++
	c.netConns[id] = netConn
	return id
}
