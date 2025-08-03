// Copyright (C) 2025 Adam Hess
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
	"context"
	"errors"
	"net"
	"sync"
	"tealfs/pkg/blockreader"
	"tealfs/pkg/blocksaver"
	"tealfs/pkg/chanutil"
	"tealfs/pkg/model"
	"tealfs/pkg/tnet"

	log "github.com/sirupsen/logrus"
)

type Conns struct {
	netConns           map[model.ConnId]tnet.RawNet
	netConnsMux        *sync.RWMutex
	nextId             model.ConnId
	acceptedConns      chan AcceptedConns
	outStatuses        chan model.NetConnectionStatus
	outReceives        chan model.ConnsMgrReceive
	OutSaveToDiskReq   chan<- blocksaver.SaveToDiskReq
	OutSaveToDiskResp  chan<- blocksaver.SaveToDiskResp
	OutGetFromDiskReq  chan<- blockreader.GetFromDiskReq
	OutGetFromDiskResp chan<- blockreader.GetFromDiskResp
	OutAddDiskReq      chan<- model.AddDiskReq
	OutIam             chan<- model.IAm
	inConnectTo        <-chan model.ConnectToNodeReq
	inSends            <-chan model.MgrConnsSend
	Address            string
	provider           ConnectionProvider
	nodeId             model.NodeId
	listener           net.Listener
	ctx                context.Context
	nodeConnMapper     model.NodeConnectionMapper
}

func NewConns(
	outStatuses chan model.NetConnectionStatus,
	outReceives chan model.ConnsMgrReceive,
	inConnectTo <-chan model.ConnectToNodeReq,
	inSends <-chan model.MgrConnsSend,
	provider ConnectionProvider,
	address string,
	nodeId model.NodeId,
	ctx context.Context,
) *Conns {
	listener, err := provider.GetListener(address)
	if err != nil {
		panic(err)
	}
	c := Conns{
		netConns:       make(map[model.ConnId]tnet.RawNet),
		netConnsMux:    &sync.RWMutex{},
		nextId:         model.ConnId(0),
		acceptedConns:  make(chan AcceptedConns),
		outStatuses:    outStatuses,
		outReceives:    outReceives,
		inConnectTo:    inConnectTo,
		inSends:        inSends,
		provider:       provider,
		nodeId:         nodeId,
		listener:       listener,
		ctx:            ctx,
		nodeConnMapper: *model.NewNodeConnectionMapper(),
	}

	go c.consumeChannels()
	go c.listen()
	go c.stopOnDone()

	return &c
}

func (c *Conns) stopOnDone() {
	<-c.ctx.Done()
	err := c.listener.Close()
	if err != nil {
		log.Warn("error closing listener")
	}
	c.netConnsMux.Lock()
	defer c.netConnsMux.Unlock()
	for connId := range c.netConns {
		conn := c.netConns[connId]
		err := conn.Close()
		if err != nil {
			log.Warn("error closing connection")
		}
	}
}

func (c *Conns) consumeChannels() {
	for {
		select {
		case <-c.ctx.Done():
			c.listener.Close()
			return
		case acceptedConn := <-c.acceptedConns:
			id := c.saveNetConn(acceptedConn.netConn)
			status := model.NetConnectionStatus{
				Type: model.Connected,
				Msg:  "Success",
				Id:   id,
			}
			chanutil.Send(c.ctx, c.outStatuses, status, "conns accepted connection sending success status")
			go c.consumeData(id)
		case connectTo := <-c.inConnectTo:
			// Todo: this needs to be non blocking
			id, err := c.connectTo(connectTo.Address)
			if err == nil {
				status := model.NetConnectionStatus{
					Type: model.Connected,
					Msg:  "Success",
					Id:   id,
				}
				chanutil.Send(c.ctx, c.outStatuses, status, "conns connected sending success status")
				go c.consumeData(id)
			}
		case sendReq := <-c.inSends:
			_, ok := c.netConns[sendReq.ConnId]
			if !ok {
				c.handleSendFailure(sendReq, errors.New("connection not found"))
			} else {
				//Todo maybe this should be async
				rawNet := c.netConns[sendReq.ConnId]
				err := rawNet.SendPayload(sendReq.Payload)
				if err != nil {
					c.handleSendFailure(sendReq, err)
				}
			}
		}
	}
}

func (c *Conns) handleSendFailure(sendReq model.MgrConnsSend, err error) {
	log.Warn("Error sending ", err)
	payload := sendReq.Payload
	switch p := payload.(type) {
	case *model.ReadRequest:
		if len(p.Ptrs) > 0 {
			ptrs := p.Ptrs[1:]
			rr := model.ReadRequest{Caller: p.Caller, Ptrs: ptrs, BlockId: p.BlockId, ReqId: p.ReqId}
			cmr := model.ConnsMgrReceive{
				ConnId:  sendReq.ConnId,
				Payload: &rr,
			}
			chanutil.Send(c.ctx, c.outReceives, cmr, "conns failed to send read request, sending new read request")
		} else {
			result := model.NewReadResultErr("no pointers in read request", p.Caller, p.ReqId, p.BlockId)
			cmr := model.ConnsMgrReceive{
				ConnId:  sendReq.ConnId,
				Payload: &result,
			}
			chanutil.Send(c.ctx, c.outReceives, cmr, "conns: failed to send read request sent failure status")
		}
	}
}

func (c *Conns) listen() {
	for {
		select {
		case <-c.ctx.Done():
			return
		default:
			conn, err := c.listener.Accept()
			if err == nil {
				incomingConnReq := AcceptedConns{netConn: conn}
				chanutil.Send(c.ctx, c.acceptedConns, incomingConnReq, "conns: accepted connection sending to acceptedConns")
			}
		}
	}
}

type AcceptedConns struct {
	netConn net.Conn
}

func (c *Conns) consumeData(conn model.ConnId) {
	ncs := model.NetConnectionStatus{
		Type: model.NotConnected,
		Msg:  "Connection closed",
		Id:   conn,
	}
	defer chanutil.Send(c.ctx, c.outStatuses, ncs, "conns connection closed sent status")

	for {
		select {
		case <-c.ctx.Done():
			return
		default:
			netConn := c.rawNetForConnId(conn)
			payload, err := netConn.ReadPayload()
			if err != nil {
				closeErr := netConn.Close()
				if closeErr != nil {
					log.Warn("Error closing connection", closeErr)
				}
				c.deleteConn(conn)
				return
			}
			switch p := (payload).(type) {
			case *blocksaver.SaveToDiskReq:
				c.OutSaveToDiskReq <- *p
			case *blocksaver.SaveToDiskResp:
				c.OutSaveToDiskResp <- *p
			case *blockreader.GetFromDiskReq:
				c.OutGetFromDiskReq <- *p
			case *blockreader.GetFromDiskResp:
				c.OutGetFromDiskResp <- *p
			case *model.AddDiskReq:
				c.OutAddDiskReq <- *p
			case *model.IAm:
				c.OutIam <- *p
				c.nodeConnMapper.SetAll(conn, p.Address, p.NodeId)
				// cmr := model.ConnsMgrReceive{
				// 	ConnId:  conn,
				// 	Payload: payload,
				// }
				// chanutil.Send(c.ctx, c.outReceives, cmr, "conns received payload sent to connsMgr "+string(c.nodeId))
			default:
				cmr := model.ConnsMgrReceive{
					ConnId:  conn,
					Payload: payload,
				}
				chanutil.Send(c.ctx, c.outReceives, cmr, "conns received payload sent to connsMgr "+string(c.nodeId))
			}
		}
	}
}

func (c *Conns) rawNetForConnId(connId model.ConnId) tnet.RawNet {
	c.netConnsMux.RLock()
	defer c.netConnsMux.RUnlock()
	return c.netConns[connId]
}

func (c *Conns) deleteConn(connId model.ConnId) {
	c.netConnsMux.Lock()
	defer c.netConnsMux.Unlock()
	delete(c.netConns, connId)
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
	c.netConnsMux.Lock()
	defer c.netConnsMux.Unlock()
	rawNet := tnet.NewRawNet(netConn)
	id := c.nextId
	c.nextId++
	c.netConns[id] = *rawNet
	return id
}
