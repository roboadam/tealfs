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
	"tealfs/pkg/webdav"

	log "github.com/sirupsen/logrus"
)

type Conns struct {
	netConns      map[model.ConnId]tnet.RawNet
	netConnsMux   *sync.RWMutex
	nextId        model.ConnId
	acceptedConns chan AcceptedConns
	// outReceives        chan model.ConnsMgrReceive
	OutSaveToDiskReq   chan<- blocksaver.SaveToDiskReq
	OutSaveToDiskResp  chan<- blocksaver.SaveToDiskResp
	OutGetFromDiskReq  chan<- blockreader.GetFromDiskReq
	OutGetFromDiskResp chan<- blockreader.GetFromDiskResp
	OutAddDiskMsg      chan<- model.AddDiskMsg
	OutDiskAddedMsg    chan<- model.DiskAddedMsg
	OutIam             chan<- model.IAm
	OutIamConnId       chan<- IamConnId
	OutSyncNodes       chan<- model.SyncNodes
	OutSendIam         chan<- model.ConnId
	OutFileBroadcasts  chan<- webdav.FileBroadcast
	inConnectTo        <-chan model.ConnectToNodeReq
	inSends            <-chan model.SendPayloadMsg
	Address            string
	provider           ConnectionProvider
	nodeId             model.NodeId
	listener           net.Listener
	ctx                context.Context
	nodeConnMapper     model.NodeConnectionMapper
}

func NewConns(
	// outReceives chan model.ConnsMgrReceive,
	inConnectTo <-chan model.ConnectToNodeReq,
	inSends <-chan model.SendPayloadMsg,
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
			c.OutSendIam <- id
			go c.consumeData(id)
		case connectTo := <-c.inConnectTo:
			id, err := c.connectTo(connectTo.Address)
			if err == nil {
				c.OutSendIam <- id
				go c.consumeData(id)
			}
		case sendReq := <-c.inSends:
			_, ok := c.netConns[sendReq.ConnId]
			if !ok {
				c.handleSendFailure(errors.New("connection not found"))
			} else {
				//Todo maybe this should be async
				rawNet := c.netConns[sendReq.ConnId]
				err := rawNet.SendPayload(sendReq.Payload)
				if err != nil {
					c.handleSendFailure(err)
				}
			}
		}
	}
}

func (c *Conns) handleSendFailure(err error) {
	log.Warn("Error sending ", err)
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
			case *model.AddDiskMsg:
				c.OutAddDiskMsg <- *p
			case *model.DiskAddedMsg:
				c.OutDiskAddedMsg <- *p
			case *model.IAm:
				c.OutIam <- *p
				c.OutIamConnId <- IamConnId{Iam: *p, ConnId: conn}
			case *model.SyncNodes:
				c.OutSyncNodes <- *p
			case *webdav.FileBroadcast:
				c.OutFileBroadcasts <- *p
			default:
				panic("Unknown payload")
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
