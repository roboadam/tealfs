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

package conns

import (
	"fmt"
	"net"
	"tealfs/pkg/model"
	"tealfs/pkg/tnet"
)

type Conns struct {
	netConns      map[model.ConnId]net.Conn
	nextId        model.ConnId
	acceptedConns chan AcceptedConns
	outStatuses   chan<- model.ConnectionStatus
	outReceives   chan<- model.ConnsMgrReceive
	inConnectTo   <-chan model.MgrConnsConnectTo
	inSends       <-chan model.MgrConnsSend
	Address       string
	provider      ConnectionProvider
}

func NewConns(outStatuses chan<- model.ConnectionStatus, outReceives chan<- model.ConnsMgrReceive, inConnectTo <-chan model.MgrConnsConnectTo, inSends <-chan model.MgrConnsSend, provider ConnectionProvider) Conns {
	listener, err := provider.GetListener("localhost:0")
	fmt.Println("LISTENING ON", listener.Addr().String())
	if err != nil {
		panic(err)
	}
	c := Conns{
		netConns:      make(map[model.ConnId]net.Conn, 3),
		nextId:        model.ConnId(0),
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
			c.outStatuses <- model.ConnectionStatus{
				Type:          model.Connected,
				Msg:           "Success",
				RemoteAddress: acceptedConn.netConn.LocalAddr().String(),
				Id:            id,
			}
			go c.consumeData(id)
		case connectTo := <-c.inConnectTo:
			// Todo: this needs to be non blocking
			id, err := c.connectTo(connectTo.Address)
			if err == nil {
				c.outStatuses <- model.ConnectionStatus{
					Type:          model.Connected,
					Msg:           "Success",
					RemoteAddress: connectTo.Address,
					Id:            id,
				}
				go c.consumeData(id)
			} else {
				c.outStatuses <- model.ConnectionStatus{
					Type:          model.NotConnected,
					Msg:           "Failure connecting",
					RemoteAddress: connectTo.Address,
					Id:            id,
				}
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
		} else {
			fmt.Println("Error accepting", err)
		}
	}
}

type AcceptedConns struct {
	netConn net.Conn
}

func (c *Conns) consumeData(conn model.ConnId) {
	for {
		netConn := c.netConns[conn]
		bytes, err := tnet.ReadPayload(netConn)
		if err != nil {
			return
		}
		payload := model.ToPayload(bytes)
		c.outReceives <- model.ConnsMgrReceive{
			ConnId:  conn,
			Payload: payload,
		}
	}
}

func (c *Conns) connectTo(address string) (model.ConnId, error) {
	netConn, err := c.provider.GetConnection(address)
	if err != nil {
		return 0, err
	}
	id := c.saveNetConn(netConn)
	return id, nil
}

func (c *Conns) saveNetConn(netConn net.Conn) model.ConnId {
	id := c.nextId
	c.nextId++
	c.netConns[id] = netConn
	return id
}
