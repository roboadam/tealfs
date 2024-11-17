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
	"fmt"
	"tealfs/pkg/disk/dist"
	"tealfs/pkg/model"
	"tealfs/pkg/set"
)

type Mgr struct {
	UiMgrConnectTos    chan model.UiMgrConnectTo
	ConnsMgrStatuses   chan model.NetConnectionStatus
	ConnsMgrReceives   chan model.ConnsMgrReceive
	DiskMgrReads       chan model.ReadResult
	DiskMgrWrites      chan model.WriteResult
	WebdavMgrGets      chan model.ReadRequest
	WebdavMgrPuts      chan model.WriteRequest
	MgrConnsConnectTos chan model.MgrConnsConnectTo
	MgrConnsSends      chan model.MgrConnsSend
	MgrDiskWrites      chan model.WriteRequest
	MgrDiskReads       chan model.ReadRequest
	MgrUiStatuses      chan model.UiConnectionStatus
	MgrWebdavGets      chan model.ReadResult
	MgrWebdavPuts      chan model.WriteResult

	nodes       set.Set[model.NodeId]
	nodeConnMap set.Bimap[model.NodeId, model.ConnId]
	NodeId      model.NodeId
	connAddress map[model.ConnId]string
	distributer dist.Distributer
	nodeAddress string
}

func NewWithChanSize(chanSize int, nodeAddress string) *Mgr {
	mgr := Mgr{
		UiMgrConnectTos:    make(chan model.UiMgrConnectTo, chanSize),
		ConnsMgrStatuses:   make(chan model.NetConnectionStatus, chanSize),
		ConnsMgrReceives:   make(chan model.ConnsMgrReceive, chanSize),
		DiskMgrWrites:      make(chan model.WriteResult),
		DiskMgrReads:       make(chan model.ReadResult, chanSize),
		WebdavMgrGets:      make(chan model.ReadRequest, chanSize),
		WebdavMgrPuts:      make(chan model.WriteRequest, chanSize),
		MgrConnsConnectTos: make(chan model.MgrConnsConnectTo, chanSize),
		MgrConnsSends:      make(chan model.MgrConnsSend, chanSize),
		MgrDiskWrites:      make(chan model.WriteRequest, chanSize),
		MgrDiskReads:       make(chan model.ReadRequest, chanSize),
		MgrUiStatuses:      make(chan model.UiConnectionStatus, chanSize),
		MgrWebdavGets:      make(chan model.ReadResult, chanSize),
		MgrWebdavPuts:      make(chan model.WriteResult, chanSize),
		nodes:              set.NewSet[model.NodeId](),
		NodeId:             model.NewNodeId(),
		connAddress:        make(map[model.ConnId]string),
		nodeConnMap:        set.NewBimap[model.NodeId, model.ConnId](),
		distributer:        dist.New(),
		nodeAddress:        nodeAddress,
	}
	mgr.distributer.SetWeight(mgr.NodeId, 1)

	return &mgr
}

func (m *Mgr) Start() {
	go m.eventLoop()
}

func (m *Mgr) eventLoop() {
	for {
		select {
		case r := <-m.UiMgrConnectTos:
			fmt.Println(m.NodeId, "Received UiMgrConnectTo")
			m.handleConnectToReq(r)
		case r := <-m.ConnsMgrStatuses:
			fmt.Println(m.NodeId, "Received ConnsMgrStatuses")
			m.handleNetConnectedStatus(r)
		case r := <-m.ConnsMgrReceives:
			fmt.Println(m.NodeId, "Received ConnsMgrReceives")
			m.handleReceives(r)
		case r := <-m.DiskMgrReads:
			fmt.Println(m.NodeId, "Received DiskMgrReads")
			m.handleDiskReads(r)
		case r := <-m.DiskMgrWrites:
			fmt.Println(m.NodeId, "Received DiskMgrWrites")
			m.handleDiskWrites(r)
		case r := <-m.WebdavMgrGets:
			fmt.Println(m.NodeId, "Received WebdavMgrGets")
			m.handleWebdavGets(r)
		case r := <-m.WebdavMgrPuts:
			fmt.Println(m.NodeId, "Received WebdavMgrPuts")
			m.handleWebdavWriteRequest(r)
		}
	}
}

func (m *Mgr) handleConnectToReq(i model.UiMgrConnectTo) {
	fmt.Println(m.NodeId, "Sending MgrConnsConnectTo")
	m.MgrConnsConnectTos <- model.MgrConnsConnectTo{Address: string(i.Address)}
}

func (m *Mgr) syncNodesPayloadToSend() model.SyncNodes {
	result := model.NewSyncNodes()
	for _, node := range m.nodes.GetValues() {
		connId, success := m.nodeConnMap.Get1(node)
		if success {
			if address, ok := m.connAddress[connId]; ok {
				result.Nodes.Add(struct {
					Node    model.NodeId
					Address string
				}{Node: node, Address: address})
			}
		}
	}
	return result
}

func (m *Mgr) handleReceives(i model.ConnsMgrReceive) {
	switch p := i.Payload.(type) {
	case *model.IAm:
		m.connAddress[i.ConnId] = p.Address
		fmt.Println(m.NodeId, "Sending MgrUiStatuses")
		m.MgrUiStatuses <- model.UiConnectionStatus{
			Type:          model.Connected,
			RemoteAddress: p.Address,
			Id:            i.ConnId,
		}
		m.addNodeToCluster(p.NodeId, i.ConnId)
		syncNodes := m.syncNodesPayloadToSend()
		for _, n := range m.nodes.GetValues() {
			connId, ok := m.nodeConnMap.Get1(n)
			if ok {
				fmt.Println(m.NodeId, "Sending MgrConnsSend SyncNodes Payload")
				m.MgrConnsSends <- model.MgrConnsSend{
					ConnId:  connId,
					Payload: &syncNodes,
				}
			}
		}
	case *model.SyncNodes:
		remoteNodes := p.GetNodes()
		localNodes := m.nodes.Clone()
		localNodes.Add(m.NodeId)
		missing := remoteNodes.Minus(&localNodes)
		for _, n := range missing.GetValues() {
			address := p.AddressForNode(n)
			fmt.Println(m.NodeId, "Sending MgrConnsConnectTo in response to SyncNodes")
			m.MgrConnsConnectTos <- model.MgrConnsConnectTo{Address: address}
		}
	case *model.WriteRequest:
		n := m.distributer.NodeIdForStoreId(p.Block.Id)
		caller, ok := m.nodeConnMap.Get2(i.ConnId)
		if !ok || caller != p.Caller {
			return
		}
		if m.NodeId == n {
			fmt.Println(m.NodeId, "Sending MgrDiskWrites")
			m.MgrDiskWrites <- *p
		} else {
			c, ok := m.nodeConnMap.Get1(n)
			if ok {
				fmt.Println(m.NodeId, "Sending MgrConnsSend in response to WriteRequest")
				m.MgrConnsSends <- model.MgrConnsSend{ConnId: c, Payload: p}
			} else {
				m.MgrDiskWrites <- *p
			}
		}
	case *model.ReadRequest:
		m.MgrDiskReads <- *p
	default:
		fmt.Println(m.NodeId, "Received unknown payload", p)
	}
}
func (m *Mgr) handleDiskWrites(r model.WriteResult) {
	if r.Caller == m.NodeId {
		fmt.Println(m.NodeId, "Sending MgrWebdavPuts WriteResult")
		m.MgrWebdavPuts <- r
	} else {
		c, ok := m.nodeConnMap.Get1(r.Caller)
		if ok {
			fmt.Println(m.NodeId, "Sending MgrConnsSend WriteResult")
			m.MgrConnsSends <- model.MgrConnsSend{ConnId: c, Payload: &r}
		} else {
			fmt.Println("Need to add to queue when reconnected")
		}
	}
}

func (m *Mgr) handleDiskReads(r model.ReadResult) {
	if r.Caller == m.NodeId {
		m.MgrWebdavGets <- r
	} else {
		c, ok := m.nodeConnMap.Get1(r.Caller)
		if ok {
			fmt.Println(m.NodeId, "Sending MgrConnsSend ReadResult")
			m.MgrConnsSends <- model.MgrConnsSend{ConnId: c, Payload: &r}
		} else {
			fmt.Println("Need to add to queue when reconnected")
		}
	}
}

func (m *Mgr) addNodeToCluster(n model.NodeId, c model.ConnId) {
	m.nodes.Add(n)
	m.nodeConnMap.Add(n, c)
	m.distributer.SetWeight(n, 1)
}

func (m *Mgr) handleNetConnectedStatus(cs model.NetConnectionStatus) {
	switch cs.Type {
	case model.Connected:
		fmt.Println(m.NodeId, "Sending MgrConnsSend Iam in response to successful connection")
		m.MgrConnsSends <- model.MgrConnsSend{
			ConnId: cs.Id,
			Payload: &model.IAm{
				NodeId:  m.NodeId,
				Address: m.nodeAddress,
			},
		}
	case model.NotConnected:
		// Todo: reflect this in the ui
		fmt.Println("Not Connected")
	}
}

func (m *Mgr) handleWebdavGets(rr model.ReadRequest) {
	n := m.distributer.NodeIdForStoreId(rr.BlockId)
	if m.NodeId == n {
		fmt.Println(m.NodeId, "Sending MgrDiskReads in response to WebdavGet ReadRequest")
		m.MgrDiskReads <- rr
	} else {
		c, ok := m.nodeConnMap.Get1(n)
		if ok {
			fmt.Println(m.NodeId, "Sending MgrConnsSend in response to WebdavGet ReadRequest")
			m.MgrConnsSends <- model.MgrConnsSend{
				ConnId:  c,
				Payload: &rr,
			}
		} else {
			fmt.Println(m.NodeId, "Sending MgrWebdavGets unsuccessful read because not connected")
			m.MgrWebdavGets <- model.ReadResult{
				Ok:      false,
				Message: "Not connected",
				Block:   model.Block{Id: rr.BlockId},
				Caller:  rr.Caller,
			}
		}
	}
}

func (m *Mgr) handleWebdavWriteRequest(w model.WriteRequest) {
	n := m.distributer.NodeIdForStoreId(w.Block.Id)
	if n == m.NodeId {
		fmt.Println(m.NodeId, "Sending MgrDiskWrites in response to webdav write request")
		m.MgrDiskWrites <- w
	} else {
		c, ok := m.nodeConnMap.Get1(n)
		if ok {
			fmt.Println(m.NodeId, "Sending MgrConnsSend in response to webdav write request")
			m.MgrConnsSends <- model.MgrConnsSend{
				ConnId:  c,
				Payload: &w,
			}
		} else {
			fmt.Println(m.NodeId, "Sending MgrDiskWrites because node is offline")
			m.MgrDiskWrites <- w
		}
	}
}
