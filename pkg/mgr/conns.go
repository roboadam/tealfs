// Copyright (C) 2024 Adam Hess
//
// This program is free software: you can redistribute it and/or modify it under
// the terms of the GNU Affero General Public License as published by the Free
// Software Foundation, version 3.
//
// This program is distributed in the hope that it will be useful, but WITHOUT
// ANY WARRANTY; without even the implied warranty of MERCHANTABILITY or
// FITNESS FOR A PARTICULAR PURPOSE. See the GNU Affero General Public License
// for more details.
//
// You should have received a copy of the GNU Affero General Public License
// along with this program. If not, see <https://www.gnu.org/licenses/>.

package mgr

import (
	"net"
	"tealfs/pkg/proto"
	"tealfs/pkg/tnet"
)

type ConnId int32
type Conns struct {
	netConns      map[ConnId]net.Conn
	nextId        ConnId
	acceptedConns chan AcceptedConns
	outStatuses   chan<- ConnsMgrStatus
	outReceives   chan<- ConnsMgrReceive
	inConnectTo   <-chan MgrConnsConnectTo
	inSends       <-chan MgrConnsSend
	Address       string
	provider      ConnectionProvider
}

func NewConns(outStatuses chan<- ConnsMgrStatus, outReceives chan<- ConnsMgrReceive, inConnectTo <-chan MgrConnsConnectTo, inSends <-chan MgrConnsSend, provider ConnectionProvider) Conns {
	listener, err := provider.GetListener("localhost:0")
	if err != nil {
		panic(err)
	}
	c := Conns{
		netConns:      make(map[ConnId]net.Conn, 3),
		nextId:        ConnId(0),
		acceptedConns: make(chan AcceptedConns),
		outStatuses:   outStatuses,
		outReceives:   outReceives,
		inConnectTo:   inConnectTo,
		inSends:       inSends,
		Address:       listener.Addr().String(),
		provider:      provider,
	}

	go c.consumeChannels()
	go c.listen(listener)

	return c
}

func (c *Conns) consumeChannels() {
	for {
		select {
		case acceptedConn := <-c.acceptedConns:
			id := c.saveNetConn(acceptedConn.netConn)
			c.outStatuses <- ConnsMgrStatus{
				Type:          Connected,
				Msg:           "Success",
				RemoteAddress: acceptedConn.netConn.LocalAddr().String(),
				Id:            id,
			}
			go c.consumeData(id)
		case connectTo := <-c.inConnectTo:
			// Todo: this needs to be non blocking
			id, err := c.connectTo(connectTo.Address)
			if err == nil {
				c.outStatuses <- ConnsMgrStatus{
					Type:          Connected,
					Msg:           "Success",
					RemoteAddress: connectTo.Address,
					Id:            id,
				}
				go c.consumeData(id)
			} else {
				// Todo
			}
		case sendReq := <-c.inSends:
			//Todo maybe this should be async
			tnet.SendPayload(c.netConns[sendReq.ConnId], sendReq.Payload.ToBytes())
		}
	}
}

func (c *Conns) listen(listener net.Listener) {
	for {
		conn, err := listener.Accept()
		if err == nil {
			incomingConnReq := AcceptedConns{netConn: conn}
			c.acceptedConns <- incomingConnReq
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

type AcceptedConns struct {
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
		c.outReceives <- ConnsMgrReceive{
			ConnId:  conn,
			Payload: payload,
		}
	}
}

func (c *Conns) connectTo(address string) (ConnId, error) {
	netConn, err := c.provider.GetConnection(address)
	if err != nil {
		return 0, err
	}
	id := c.saveNetConn(netConn)
	return id, nil
}

func (c *Conns) saveNetConn(netConn net.Conn) ConnId {
	id := c.nextId
	c.nextId++
	c.netConns[id] = netConn
	return id
}
