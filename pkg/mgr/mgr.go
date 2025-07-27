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
)

type Mgr struct {
	ConnectToNodeReqs       chan<- model.ConnectToNodeReq
	UiMgrDisk               chan model.AddNewDiskReq
	ConnsMgrStatuses        chan model.NetConnectionStatus
	ConnsMgrReceives        chan model.ConnsMgrReceive
	DiskMgrReads            chan model.ReadResult
	WebdavMgrBroadcast      chan model.Broadcast
	MgrConnsSends           chan model.MgrConnsSend
	MgrUiConnectionStatuses chan model.UiConnectionStatus
	MgrUiDiskStatuses       chan model.UiDiskStatus
	MgrWebdavBroadcast      chan model.Broadcast
	CustodianCommands       chan<- custodian.Command

	nodeConnMapper *model.NodeConnectionMapper

	NodeId             model.NodeId
	nodeAddress        string
	savePath           string
	fileOps            disk.FileOps
	pendingBlockWrites pendingBlockWrites
	freeBytes          uint32
	ctx                context.Context
}

func New(
	chanSize int,
	nodeAddress string,
	globalPath string,
	fileOps disk.FileOps,
	freeBytes uint32,
	nodeConnMapper *model.NodeConnectionMapper,
	ctx context.Context,
) *Mgr {
	nodeId, err := readNodeId(globalPath, fileOps)
	if err != nil {
		panic(err)
	}

	mgr := Mgr{
		UiMgrDisk:               make(chan model.AddNewDiskReq, chanSize),
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
		pendingBlockWrites:      newPendingBlockWrites(),
		freeBytes:               freeBytes,
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

// func (m *Mgr) createLocalDisk(id model.DiskId, path string) bool {
// 	for _, disk := range m.Disks.GetValues() {
// 		if disk.Id() == id {
// 			return false
// 		}
// 	}
// 	p := disk.NewPath(path, m.fileOps)
// 	d := disk.New(p, m.NodeId, id, m.ctx)
// 	m.Disks.Add(d)
// 	for _, dChan := range m.AddedDisk {
// 		dChan <- &d
// 	}
// 	return true
// }

// func (m *Mgr) markLocalDiskAvailable(id model.DiskId, path string, size int) {
// 	m.MirrorDistributer.SetWeight(m.NodeId, id, size)
// 	status := model.UiDiskStatus{
// 		Localness:     model.Local,
// 		Availableness: model.Available,
// 		Node:          m.NodeId,
// 		Id:            id,
// 		Path:          path,
// 	}
// 	chanutil.Send(m.ctx, m.MgrUiDiskStatuses, status, "mgr: local disk available")
// }

func (m *Mgr) markRemoteDiskUnknown(id model.DiskId, node model.NodeId, path string) {
	status := model.UiDiskStatus{
		Localness:     model.Remote,
		Availableness: model.Unknown,
		Node:          node,
		Id:            id,
		Path:          path,
	}
	chanutil.Send(m.ctx, m.MgrUiDiskStatuses, status, "mgr: remote disk unknown")
}

func (m *Mgr) loadSettings() error {
	data, err := m.fileOps.ReadFile(filepath.Join(m.savePath, "cluster.json"))
	if err != nil && !errors.Is(err, fs.ErrNotExist) {
		return err
	}
	if len(data) > 0 {
		mapper, err := model.NodeConnectionMapperUnmarshal(data)
		if err != nil {
			return err
		}
		mapper.UnsetConnections()
		m.nodeConnMapper = mapper
	}

	// data, err = m.fileOps.ReadFile(filepath.Join(m.savePath, "disks.json"))
	// if err == nil {
	// 	err = json.Unmarshal(data, &m.DiskIds)
	// 	if err != nil {
	// 		return err
	// 	}
	// } else if errors.Is(err, fs.ErrNotExist) {
	// 	m.DiskIds = []model.DiskIdPath{}
	// } else {
	// 	return err
	// }

	// m.syncDisksAndIds()

	return nil
}

func (m *Mgr) saveSettings() error {
	// data, err := json.Marshal(m.DiskIds)
	// if err != nil {
	// 	return err
	// }

	// err = m.fileOps.WriteFile(filepath.Join(m.savePath, "disks.json"), data)
	// if err != nil {
	// 	return err
	// }

	data, err := m.nodeConnMapper.Marshal()
	if err != nil {
		return err
	}

	err = m.fileOps.WriteFile(filepath.Join(m.savePath, "cluster.json"), data)
	if err != nil {
		return err
	}

	return nil
}

func (m *Mgr) eventLoop() {
	for {
		select {
		case <-m.ctx.Done():
			return
		// case r := <-m.UiMgrDisk:
		// 	m.handleAddDiskReq(r)
		case r := <-m.ConnsMgrStatuses:
			m.handleNetConnectedStatus(r)
		case r := <-m.ConnsMgrReceives:
			m.handleReceives(r)
		case r := <-m.WebdavMgrBroadcast:
			m.handleWebdavMgrBroadcast(r)
		}
	}
}

// func (m *Mgr) syncDisksAndIds() {
// 	for _, diskIdPath := range m.DiskIds {
// 		if diskIdPath.Node == m.NodeId {
// 			m.createDiskChannels(diskIdPath.Id)
// 			m.createLocalDisk(diskIdPath.Id, diskIdPath.Path)
// 			m.markLocalDiskAvailable(diskIdPath.Id, diskIdPath.Path, 1)
// 		} else {
// 			m.markRemoteDiskUnknown(diskIdPath.Id, diskIdPath.Node, diskIdPath.Path)
// 		}
// 	}
// }

// func (m *Mgr) handleAddDiskReq(i model.AddDiskReq) {
// 	if i.Node == m.NodeId {
// 		id := model.DiskId(uuid.New().String())
// 		m.DiskIds = append(m.DiskIds, model.DiskIdPath{Id: id, Path: i.Path, Node: m.NodeId})
// 		m.syncDisksAndIds()
// 		err := m.saveSettings()
// 		if err != nil {
// 			panic("error saving disk settings")
// 		}

// 		connections := m.nodeConnMapper.Connections()
// 		for _, conn := range connections.GetValues() {
// 			iam := model.NewIam(m.NodeId, m.DiskIds, m.nodeAddress, m.freeBytes)
// 			mcs := model.MgrConnsSend{
// 				ConnId:  conn,
// 				Payload: &iam,
// 			}
// 			chanutil.Send(m.ctx, m.MgrConnsSends, mcs, "mgr: handleDiskReq: added disk")
// 		}
// 	} else {
// 		if conn, exists := m.nodeConnMapper.ConnForNode(i.Node); exists {
// 			chanutil.Send(m.ctx, m.MgrConnsSends, model.MgrConnsSend{ConnId: conn, Payload: &i}, "send add disk to node")
// 		}
// 	}
// }

func (m *Mgr) syncNodesPayloadToSend() model.SyncNodes {
	result := model.NewSyncNodes()
	addressesAndNodes := m.nodeConnMapper.AddressesAndNodes()
	for _, an := range addressesAndNodes.GetValues() {
		result.Nodes.Add(struct {
			Node    model.NodeId
			Address string
		}{Node: an.NodeId, Address: an.Address})
	}
	return result
}

func (m *Mgr) handleReceives(i model.ConnsMgrReceive) {
	switch p := i.Payload.(type) {
	case *model.IAm:
		m.nodeConnMapper.SetAll(i.ConnId, p.Address, p.NodeId)
		status := model.UiConnectionStatus{
			Type:          model.Connected,
			RemoteAddress: p.Address,
			Id:            p.NodeId,
		}
		chanutil.Send(m.ctx, m.MgrUiConnectionStatuses, status, "mgr: handleReceives: ui status")
		_ = m.addNodeToCluster(*p, i.ConnId)
		for _, d := range p.Disks {
			diskStatus := model.UiDiskStatus{
				Localness:     model.Remote,
				Availableness: model.Available,
				Node:          p.NodeId,
				Id:            d.Id,
				Path:          d.Path,
			}
			chanutil.Send(m.ctx, m.MgrUiDiskStatuses, diskStatus, "mgr: handleReceives: ui disk status")
		}
		syncNodes := m.syncNodesPayloadToSend()
		connections := m.nodeConnMapper.Connections()
		for _, connId := range connections.GetValues() {
			mcs := model.MgrConnsSend{
				ConnId:  connId,
				Payload: &syncNodes,
			}
			chanutil.Send(m.ctx, m.MgrConnsSends, mcs, "mgr: handleReceives: sync nodes")
		}
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
	// case *model.AddDiskReq:
	// 	m.handleAddDiskReq(*p)
	default:
		panic("Received unknown payload")
	}
}

func (m *Mgr) addNodeToCluster(iam model.IAm, c model.ConnId) error {
	m.nodeConnMapper.SetAll(c, iam.Address, iam.NodeId)
	err := m.saveSettings()
	if err != nil {
		return err
	}
	// for _, disk := range iam.Disks {
	// 	m.MirrorDistributer.SetWeight(iam.NodeId, disk.Id, int(iam.FreeBytes))
	// }
	return nil
}

func (m *Mgr) handleNetConnectedStatus(cs model.NetConnectionStatus) {
	switch cs.Type {
	case model.Connected:
		iam := model.NewIam(m.NodeId, m.DiskIds, m.nodeAddress, m.freeBytes)
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
