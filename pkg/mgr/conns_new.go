package mgr

import "net"

type ConnNewId int32
type ConnsNew struct {
	netConns map[ConnNewId]net.Conn
	nextId   ConnNewId
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
	resp    chan<- IncomingConnResp
}
type IncomingConnResp struct {
	Success      bool
	Id           ConnNewId
	ErrorMessage string
}

func (c *ConnsNew) HandleIncoming(req IncomingConnReq) {
	id := c.saveNetConn(req.netConn)
	req.resp <- IncomingConnResp{Success: true, Id: id}
}

func (c *ConnsNew) ConnectTo(req ConnectToReq) {
	netConn, err := net.Dial("tcp", req.address)
	if err != nil {
		req.resp <- ConnectToResp{Success: false}
	}
	id := c.saveNetConn(netConn)
	req.resp <- ConnectToResp{Success: true, Id: id}
}

func (c *ConnsNew) saveNetConn(netConn net.Conn) ConnNewId {
	id := c.nextId
	c.nextId++
	c.netConns[id] = netConn
	return id
}
