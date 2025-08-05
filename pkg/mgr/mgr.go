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

package mgr

import (
	"context"
	"errors"
	"io/fs"
	"os"
	"path/filepath"
	"tealfs/pkg/chanutil"
	"tealfs/pkg/custodian"
	"tealfs/pkg/disk"
	"tealfs/pkg/model"
	"tealfs/pkg/set"
)

type Mgr struct {
	ConnectToNodeReqs       chan<- model.ConnectToNodeReq
	ConnsMgrStatuses        chan model.NetConnectionStatus
	ConnsMgrReceives        chan model.ConnsMgrReceive
	DiskMgrReads            chan model.ReadResult
	WebdavMgrBroadcast      chan model.Broadcast
	MgrConnsSends           chan model.MgrConnsSend
	MgrUiConnectionStatuses chan model.UiConnectionStatus
	MgrUiDiskStatuses       chan model.UiDiskStatus
	MgrWebdavBroadcast      chan model.Broadcast
	CustodianCommands       chan<- custodian.Command
	AllDiskIds              *set.Set[model.AddDiskReq]

	nodeConnMapper *model.NodeConnectionMapper

	NodeId      model.NodeId
	nodeAddress string
	savePath    string
	fileOps     disk.FileOps
	ctx         context.Context
}

func New(
	chanSize int,
	nodeAddress string,
	globalPath string,
	fileOps disk.FileOps,
	nodeConnMapper *model.NodeConnectionMapper,
	ctx context.Context,
) *Mgr {
	nodeId, err := readNodeId(globalPath, fileOps)
	if err != nil {
		panic(err)
	}

	mgr := Mgr{
		ConnsMgrStatuses:        make(chan model.NetConnectionStatus, chanSize),
		ConnsMgrReceives:        make(chan model.ConnsMgrReceive, chanSize),
		DiskMgrReads:            make(chan model.ReadResult, chanSize),
		WebdavMgrBroadcast:      make(chan model.Broadcast, chanSize),
		MgrConnsSends:           make(chan model.MgrConnsSend, chanSize),
		MgrUiConnectionStatuses: make(chan model.UiConnectionStatus, chanSize),
		MgrUiDiskStatuses:       make(chan model.UiDiskStatus, chanSize),
		MgrWebdavBroadcast:      make(chan model.Broadcast, chanSize),
		nodeConnMapper:          nodeConnMapper,
		NodeId:                  nodeId,
		nodeAddress:             nodeAddress,
		savePath:                globalPath,
		fileOps:                 fileOps,
		ctx:                     ctx,
	}

	return &mgr
}

func readNodeId(savePath string, fileOps disk.FileOps) (model.NodeId, error) {
	data, err := fileOps.ReadFile(filepath.Join(savePath, "node_id"))
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			nodeId := model.NewNodeId()
			err = fileOps.WriteFile(filepath.Join(savePath, "node_id"), []byte(nodeId))
			if err != nil {
				return "", err
			}
			return nodeId, nil
		}
		return "", err
	}
	return model.NodeId(data), nil
}

func (m *Mgr) Start() {
	go func() {
		err := m.loadSettings()
		if err != nil {
			panic("Unable to load settings " + err.Error())
		}
		go m.eventLoop()
		m.connectToUnconnected()
	}()
}

func (m *Mgr) connectToUnconnected() {
	addresses := m.nodeConnMapper.AddressesWithoutConnections()
	for _, address := range addresses.GetValues() {
		m.ConnectToNodeReqs <- model.ConnectToNodeReq{Address: address}
	}
}

func (m *Mgr) eventLoop() {
	for {
		select {
		case <-m.ctx.Done():
			return
		case r := <-m.ConnsMgrStatuses:
			m.handleNetConnectedStatus(r)
		case r := <-m.ConnsMgrReceives:
			m.handleReceives(r)
		case r := <-m.WebdavMgrBroadcast:
			m.handleWebdavMgrBroadcast(r)
		}
	}
}

func (m *Mgr) handleReceives(i model.ConnsMgrReceive) {
	switch p := i.Payload.(type) {
	case *model.SyncNodes:
		remoteNodes := p.GetNodes()
		localNodes := m.nodeConnMapper.Nodes()
		localNodes.Add(m.NodeId)
		missing := remoteNodes.Minus(&localNodes)
		for _, n := range missing.GetValues() {
			address := p.AddressForNode(n)
			mct := model.ConnectToNodeReq{Address: address}
			chanutil.Send(m.ctx, m.ConnectToNodeReqs, mct, "mgr: handleReceives: connect to")

		}
	case *model.Broadcast:
		chanutil.Send(m.ctx, m.MgrWebdavBroadcast, *p, "mgr: handleReceives: forward broadcast to webdav")
	default:
		panic("Received unknown payload")
	}
}

func (m *Mgr) handleNetConnectedStatus(cs model.NetConnectionStatus) {
	switch cs.Type {
	case model.Connected:
		iam := model.NewIam(m.NodeId, m.AllDiskIds.GetValues(), m.nodeAddress)
		mcs := model.MgrConnsSend{
			ConnId:  cs.Id,
			Payload: &iam,
		}
		chanutil.Send(m.ctx, m.MgrConnsSends, mcs, "mgr: handleNetConnectedStatus: connected")
	case model.NotConnected:
		address, _ := m.nodeConnMapper.AddressForConn(cs.Id)
		id, _ := m.nodeConnMapper.NodeForConn(cs.Id)
		m.nodeConnMapper.RemoveConn(cs.Id)
		m.MgrUiConnectionStatuses <- model.UiConnectionStatus{
			Type:          model.NotConnected,
			RemoteAddress: address,
			Msg:           "Disconnected",
			Id:            id,
		}
		// Todo: Need to periodically try to reconnect to unconnected nodes
	}
}

func (m *Mgr) handleWebdavMgrBroadcast(b model.Broadcast) {
	connections := m.nodeConnMapper.Connections()
	for _, connId := range connections.GetValues() {
		chanutil.Send(m.ctx, m.MgrConnsSends, model.MgrConnsSend{ConnId: connId, Payload: &b}, "Broadcasting")
	}
}
