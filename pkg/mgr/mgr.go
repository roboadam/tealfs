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
	"tealfs/pkg/model"
	"tealfs/pkg/nodes"
	"tealfs/pkg/proto"
	"tealfs/pkg/set"
	"tealfs/pkg/store"
	"tealfs/pkg/store/dist"
)

type Mgr struct {
	UiMgrConnectTos    chan model.UiMgrConnectTo
	ConnsMgrStatuses   chan model.ConnsMgrStatus
	ConnsMgrReceives   chan model.ConnsMgrReceive
	DiskMgrReads       chan proto.ReadResult
	DiskMgrWrites      chan proto.WriteResult
	WebdavMgrGets      chan proto.ReadRequest
	WebdavMgrPuts      chan store.Block
	MgrConnsConnectTos chan model.MgrConnsConnectTo
	MgrConnsSends      chan model.MgrConnsSend
	MgrDiskWrites      chan store.Block
	MgrDiskReads       chan proto.ReadRequest
	MgrWebdavGets      chan proto.ReadResult
	MgrWebdavPuts      chan proto.WriteResult

	nodes       set.Set[nodes.Id]
	nodeConnMap set.Bimap[nodes.Id, model.ConnId]
	nodeId      nodes.Id
	connAddress map[model.ConnId]string
	distributer dist.Distributer
}

func NewWithChanSize(chanSize int) Mgr {
	mgr := Mgr{
		UiMgrConnectTos:    make(chan model.UiMgrConnectTo, chanSize),
		ConnsMgrStatuses:   make(chan model.ConnsMgrStatus, chanSize),
		ConnsMgrReceives:   make(chan model.ConnsMgrReceive, chanSize),
		DiskMgrReads:       make(chan proto.ReadResult, chanSize),
		WebdavMgrGets:      make(chan proto.ReadRequest, chanSize),
		WebdavMgrPuts:      make(chan store.Block, chanSize),
		MgrConnsConnectTos: make(chan model.MgrConnsConnectTo, chanSize),
		MgrConnsSends:      make(chan model.MgrConnsSend, chanSize),
		MgrDiskWrites:      make(chan store.Block, chanSize),
		MgrDiskReads:       make(chan proto.ReadRequest, chanSize),
		MgrWebdavGets:      make(chan proto.ReadResult, chanSize),
		MgrWebdavPuts:      make(chan proto.WriteResult, chanSize),
		nodes:              set.NewSet[nodes.Id](),
		nodeId:             nodes.NewNodeId(),
		connAddress:        make(map[model.ConnId]string),
		nodeConnMap:        set.NewBimap[nodes.Id, model.ConnId](),
		distributer:        dist.New(),
	}
	mgr.distributer.SetWeight(mgr.nodeId, 1)

	return mgr
}

func (m *Mgr) Start() {
	go m.eventLoop()
}

func (m *Mgr) eventLoop() {
	for {
		select {
		case r := <-m.UiMgrConnectTos:
			m.handleConnectToReq(r)
		case r := <-m.ConnsMgrStatuses:
			m.handleConnectedStatus(r)
		case r := <-m.ConnsMgrReceives:
			m.handleReceives(r)
		case r := <-m.DiskMgrReads:
			m.handleDiskReads(r)
		case r := <-m.WebdavMgrGets:
			m.handleWebdavGets(r)
		case r := <-m.WebdavMgrPuts:
			m.handlePuts(r)
		}
	}
}

func (m *Mgr) handleConnectToReq(i model.UiMgrConnectTo) {
	m.MgrConnsConnectTos <- model.MgrConnsConnectTo{Address: string(i.Address)}
}

func (m *Mgr) syncNodesPayloadToSend() proto.SyncNodes {
	result := proto.NewSyncNodes()
	for _, node := range m.nodes.GetValues() {
		connId, success := m.nodeConnMap.Get1(node)
		if success {
			if address, ok := m.connAddress[connId]; ok {
				result.Nodes.Add(struct {
					Node    nodes.Id
					Address string
				}{Node: node, Address: address})
			}
		}
	}
	return result
}

func (m *Mgr) handleReceives(i model.ConnsMgrReceive) {
	switch p := i.Payload.(type) {
	case *proto.IAm:
		m.addNodeToCluster(p.NodeId, i.ConnId)
		syncNodes := m.syncNodesPayloadToSend()
		for _, n := range m.nodes.GetValues() {
			connId, ok := m.nodeConnMap.Get1(n)
			if ok {
				m.MgrConnsSends <- model.MgrConnsSend{
					ConnId:  connId,
					Payload: &syncNodes,
				}
			}
		}
	case *proto.SyncNodes:
		remoteNodes := p.GetNodes()
		localNodes := m.nodes.Clone()
		localNodes.Add(m.nodeId)
		missing := remoteNodes.Minus(&m.nodes)
		for _, n := range missing.GetValues() {
			address := p.AddressForNode(n)
			m.MgrConnsConnectTos <- model.MgrConnsConnectTo{Address: address}
		}
	case *proto.SaveData:
		n := m.distributer.NodeIdForStoreId(p.Block.Id)
		if m.nodeId == n {
			m.MgrDiskWrites <- p.Block
		} else {
			c, ok := m.nodeConnMap.Get1(n)
			if ok {
				m.MgrConnsSends <- model.MgrConnsSend{ConnId: c, Payload: p}
			} else {
				m.MgrDiskWrites <- p.Block
			}
		}
	}
}

func (m *Mgr) handleDiskReads(r proto.ReadResult) {
	if r.Caller == m.nodeId {
		m.MgrWebdavGets <- r
	} else {
		c, ok := m.nodeConnMap.Get1(r.Caller)
		if ok {
			m.MgrConnsSends <- model.MgrConnsSend{ConnId: c, Payload: &r}
		} else {
			panic("Oh no")
			// Todo: need a ticket to create queuing for offline nodes
		}
	}
}

func (m *Mgr) addNodeToCluster(n nodes.Id, c model.ConnId) {
	m.nodes.Add(n)
	m.nodeConnMap.Add(n, c)
	m.distributer.SetWeight(n, 1)
}

func (m *Mgr) handleConnectedStatus(cs model.ConnsMgrStatus) {
	switch cs.Type {
	case model.Connected:
		m.connAddress[cs.Id] = cs.RemoteAddress
		m.MgrConnsSends <- model.MgrConnsSend{
			ConnId:  cs.Id,
			Payload: &proto.IAm{NodeId: m.nodeId},
		}
	case model.NotConnected:
		// Todo: reflect this in the ui
		println("Not Connected")
	}
}

func (m *Mgr) handleWebdavGets(rr proto.ReadRequest) {
	n := m.distributer.NodeIdForStoreId(rr.BlockId)
	if m.nodeId == n {
		m.MgrDiskReads <- rr
	} else {
		c, ok := m.nodeConnMap.Get1(n)
		if ok {
			m.MgrConnsSends <- model.MgrConnsSend{
				ConnId:  c,
				Payload: &rr,
			}
		} else {
			m.MgrWebdavGets <- proto.ReadResult{
				Ok:      false,
				Message: "Not connected",
				Block:   store.Block{Id: rr.BlockId},
				Caller:  rr.Caller,
			}
		}
	}
}

func (m *Mgr) handlePuts(w store.Block) {
	n := m.distributer.NodeIdForStoreId(w.Id)
	if n == m.nodeId {
		m.MgrDiskWrites <- w
	} else {
		c, ok := m.nodeConnMap.Get1(n)
		if ok {
			m.MgrConnsSends <- model.MgrConnsSend{
				ConnId: c,
				Payload: &proto.WriteRequest{
					Caller: m.nodeId,
					Block:  w,
				},
			}
		} else {
			panic("no connection!")
			// Todo handle no connections here
		}
	}
}
