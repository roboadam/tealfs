// Copyright (C) 2026 Adam Hess
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
	"tealfs/pkg/blocksaver"
	"tealfs/pkg/datalayer"
	"tealfs/pkg/disk"
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

	OutSaveToDiskResp     chan<- blocksaver.SaveToDiskResp
	OutAddDiskMsg         chan<- model.AddDiskMsg
	OutDiskAddedMsg       chan<- model.DiskAddedMsg
	OutIam                chan<- model.IAm
	OutFileBroadcasts     chan<- webdav.FileBroadcast
	OutDataForSaveRequest chan<- datalayer.DataForSaveRequest
	OutSaveRequest        chan<- datalayer.SaveRequest
	OutFetchBlockResp     chan<- model.FetchBlockResp

	inConnectTo chan model.ConnectToNodeReq

	InPayload <-chan model.Payload2

	StateHandler         *datalayer.StateHandler
	Address              string
	DiskManager          *disk.DiskManagerSvc
	DeleteRequestHandler *datalayer.DeleteRequestHandler
	LocalBlockSaver      *blocksaver.LocalBlockSaver
	IamGenerator         *IamSender

	provider       ConnectionProvider
	nodeId         model.NodeId
	listener       net.Listener
	ctx            context.Context
	NodeConnMapper *model.NodeConnectionMapper
}

func NewConns(
	inConnectTo chan model.ConnectToNodeReq,
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
		netConns:      make(map[model.ConnId]tnet.RawNet),
		netConnsMux:   &sync.RWMutex{},
		nextId:        model.ConnId(0),
		acceptedConns: make(chan AcceptedConns),
		inConnectTo:   inConnectTo,
		provider:      provider,
		nodeId:        nodeId,
		listener:      listener,
		ctx:           ctx,
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
		log.Warn("error closing listener ", err)
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
			iam := c.IamGenerator.generateIam()
			err := c.sendPayloadOnConn(iam, id)
			if err != nil {
				log.Panic("no connection")
			}
			go c.consumeData(id)
		case connectTo := <-c.inConnectTo:
			id, err := c.connectTo(connectTo.Address)
			if err == nil {
				iam := c.IamGenerator.generateIam()
				err := c.sendPayloadOnConn(iam, id)
				if err != nil {
					log.Panic("no connection")
				}
				go c.consumeData(id)
			}
		case payload := <-c.InPayload:
			c.sendPayload(payload)
		}
	}
}

func (c *Conns) handleIncomingPayload(payload model.Payload2) {
	switch p := payload.(type) {
	case *model.FetchBlockReq:
		c.handleFetchBlockReq(p)
	case *model.FetchBlockCmd:
		c.handleFetchBlockCmd(p)
	case *model.FetchBlockResp:
		c.OutFetchBlockResp <- *p
	case *datalayer.DeleteRequest:
		go c.DeleteRequestHandler.HandleDeleteRequest(p)
	case *blocksaver.SaveToDiskReq:
		go c.LocalBlockSaver.Save(*p)
	case *blocksaver.SaveToDiskResp:
		c.OutSaveToDiskResp <- *p
	case *model.AddDiskMsg:
		c.OutAddDiskMsg <- *p
	case *model.DiskAddedMsg:
		c.OutDiskAddedMsg <- *p
	case *model.IAm:
		c.OutIam <- *p
	case *webdav.FileBroadcast:
		c.OutFileBroadcasts <- *p
	case *datalayer.DeletedParams:
		c.StateHandler.Deleted(p.B, p.D)
	case *datalayer.SavedParams:
		c.StateHandler.Saved(p.B, p.D)
	case *datalayer.SetDiskSpaceParams:
		c.StateHandler.SetDiskSpace(p.D, p.Space)
	case *datalayer.SaveRequest:
		c.OutSaveRequest <- *p
	case *datalayer.DataForSaveRequest:
		c.OutDataForSaveRequest <- *p
	default:
		log.Panic("Unknown data type")
	}
}

func (c *Conns) handleFetchBlockCmd(p *model.FetchBlockCmd) {
	for i, source := range p.Sources {
		if source.NodeId != c.nodeId {
			p.Sources = p.Sources[i:]
			c.sendPayload(p)
			return
		}
		if data, ok := c.DiskManager.Get(p.BlockId, source.DiskId); ok {
			resp := model.FetchBlockResp{
				Caller: p.Caller,
				Block: model.Block{
					Id:   p.BlockId,
					Data: data,
				},
				Id:      p.Id,
				Success: true,
			}
			c.sendPayload(&resp)
			return
		}
	}
}

func (c *Conns) handleFetchBlockReq(p *model.FetchBlockReq) {
	cmd, err := c.StateHandler.FetchBlockReqToCmd(p)
	if errors.Is(err, datalayer.NotMainNodeErr{}) {
		c.sendPayload(p)
	} else if err == nil {
		c.sendPayload(cmd)
	}
}

func (c *Conns) sendPayload(p model.Payload2) error {
	dest := p.Destination()
	if dest == "" {
		dest = c.NodeConnMapper.MainNode(c.nodeId)
	}

	if dest == c.nodeId {
		c.handleIncomingPayload(p)
		return nil
	}

	connId, ok := c.NodeConnMapper.ConnForNode(dest)
	if !ok {
		return errors.New("No connection to that node")
	}
	rawNet := c.netConns[connId]
	err := rawNet.SendPayload2(p)
	return err
}

func (c *Conns) sendPayloadOnConn(p model.Payload2, connId model.ConnId) error {
	rawNet := c.netConns[connId]
	return rawNet.SendPayload2(p)
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
				c.acceptedConns <- incomingConnReq
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

			payload2, err := netConn.ReadPayload2()
			if err != nil {
				log.Errorf("Error reading payload %s", err)
				return
			}
			if iam, ok := payload2.(*model.IAm); ok {
				c.NodeConnMapper.SetAll(conn, iam.Node.Address, iam.Node.NodeId)
			}
			c.handleIncomingPayload(payload2)
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
