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
	"encoding/json"
	"errors"
	"io/fs"
	"os"
	"path/filepath"
	"tealfs/pkg/chanutil"
	"tealfs/pkg/disk"
	"tealfs/pkg/disk/dist"
	"tealfs/pkg/model"
	"tealfs/pkg/set"

	"github.com/google/uuid"
	log "github.com/sirupsen/logrus"
)

type Mgr struct {
	UiMgrConnectTos         chan model.UiMgrConnectTo
	UiMgrDisk               chan model.AddDiskReq
	ConnsMgrStatuses        chan model.NetConnectionStatus
	ConnsMgrReceives        chan model.ConnsMgrReceive
	DiskMgrReads            chan model.ReadResult
	DiskMgrWrites           chan model.WriteResult
	WebdavMgrGets           chan model.GetBlockReq
	WebdavMgrPuts           chan model.PutBlockReq
	WebdavMgrBroadcast      chan model.Broadcast
	MgrConnsConnectTos      chan model.MgrConnsConnectTo
	MgrConnsSends           chan model.MgrConnsSend
	MgrDiskWrites           map[model.DiskId]chan model.WriteRequest
	MgrDiskReads            map[model.DiskId]chan model.ReadRequest
	MgrUiConnectionStatuses chan model.UiConnectionStatus
	MgrUiDiskStatuses       chan model.UiDiskStatus
	MgrWebdavGets           chan model.GetBlockResp
	MgrWebdavPuts           chan model.PutBlockResp
	MgrWebdavBroadcast      chan model.Broadcast

	nodesAddressMap    map[model.NodeId]string
	nodeConnMap        set.Bimap[model.NodeId, model.ConnId]
	NodeId             model.NodeId
	connAddress        set.Bimap[model.ConnId, string]
	mirrorDistributer  dist.MirrorDistributer
	blockType          model.BlockType
	nodeAddress        string
	savePath           string
	fileOps            disk.FileOps
	pendingBlockWrites pendingBlockWrites
	freeBytes          uint32
	DiskIds            []model.DiskIdPath
	disks              []disk.Disk
	ctx                context.Context
}

func NewWithChanSize(
	chanSize int,
	nodeAddress string,
	globalPath string,
	fileOps disk.FileOps,
	blockType model.BlockType,
	freeBytes uint32,
	ctx context.Context,
) *Mgr {
	nodeId, err := readNodeId(globalPath, fileOps)
	if err != nil {
		panic(err)
	}

	mgr := Mgr{
		UiMgrConnectTos:         make(chan model.UiMgrConnectTo, chanSize),
		UiMgrDisk:               make(chan model.AddDiskReq),
		ConnsMgrStatuses:        make(chan model.NetConnectionStatus, chanSize),
		ConnsMgrReceives:        make(chan model.ConnsMgrReceive, chanSize),
		DiskMgrWrites:           make(chan model.WriteResult),
		DiskMgrReads:            make(chan model.ReadResult, chanSize),
		WebdavMgrGets:           make(chan model.GetBlockReq, chanSize),
		WebdavMgrPuts:           make(chan model.PutBlockReq, chanSize),
		WebdavMgrBroadcast:      make(chan model.Broadcast, chanSize),
		MgrConnsConnectTos:      make(chan model.MgrConnsConnectTo, chanSize),
		MgrConnsSends:           make(chan model.MgrConnsSend, chanSize),
		MgrDiskWrites:           make(map[model.DiskId]chan model.WriteRequest),
		MgrDiskReads:            make(map[model.DiskId]chan model.ReadRequest),
		MgrUiConnectionStatuses: make(chan model.UiConnectionStatus, chanSize),
		MgrUiDiskStatuses:       make(chan model.UiDiskStatus, chanSize),
		MgrWebdavGets:           make(chan model.GetBlockResp, chanSize),
		MgrWebdavPuts:           make(chan model.PutBlockResp, chanSize),
		MgrWebdavBroadcast:      make(chan model.Broadcast, chanSize),
		nodesAddressMap:         make(map[model.NodeId]string),
		NodeId:                  nodeId,
		connAddress:             set.NewBimap[model.ConnId, string](),
		nodeConnMap:             set.NewBimap[model.NodeId, model.ConnId](),
		mirrorDistributer:       dist.NewMirrorDistributer(),
		blockType:               blockType,
		nodeAddress:             nodeAddress,
		savePath:                globalPath,
		fileOps:                 fileOps,
		pendingBlockWrites:      newPendingBlockWrites(),
		freeBytes:               freeBytes,
		DiskIds:                 []model.DiskIdPath{},
		disks:                   []disk.Disk{},
		ctx:                     ctx,
	}

	go func() {
		err = mgr.loadSettings()
		if err != nil {
			panic("Unable to load settings " + err.Error())
		}
		mgr.start()
	}()

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

func (m *Mgr) start() {
	go m.eventLoop()
	for nodeId, address := range m.nodesAddressMap {
		if nodeId != m.NodeId {
			chanutil.Send(m.ctx, m.UiMgrConnectTos, model.UiMgrConnectTo{Address: address}, "mgr: init connect to")
		}
	}
}

func (m *Mgr) createDiskChannels(diskId model.DiskId) {
	if _, ok := m.MgrDiskReads[diskId]; !ok {
		m.MgrDiskReads[diskId] = make(chan model.ReadRequest)
	}
	if _, ok := m.MgrDiskWrites[diskId]; !ok {
		m.MgrDiskWrites[diskId] = make(chan model.WriteRequest)
	}
}

func (m *Mgr) createLocalDisk(id model.DiskId, path string) bool {
	for _, disk := range m.disks {
		if disk.Id() == id {
			return false
		}
	}
	p := disk.NewPath(path, m.fileOps)
	d := disk.New(
		p,
		m.NodeId,
		id,
		m.MgrDiskWrites[id],
		m.MgrDiskReads[id],
		m.DiskMgrWrites,
		m.DiskMgrReads,
		m.ctx,
	)
	m.disks = append(m.disks, d)
	return true
}

func (m *Mgr) markLocalDiskAvailable(id model.DiskId, path string, size int) {
	m.mirrorDistributer.SetWeight(m.NodeId, id, size)
	status := model.UiDiskStatus{
		Localness:     model.Local,
		Availableness: model.Available,
		Node:          m.NodeId,
		Id:            id,
		Path:          path,
	}
	chanutil.Send(m.ctx, m.MgrUiDiskStatuses, status, "mgr: local disk available")
}

func (m *Mgr) markRemoteDiskUnknown(id model.DiskId, node model.NodeId, path string, size int) {
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
		err = json.Unmarshal(data, &m.nodesAddressMap)
		if err != nil {
			return err
		}
	}

	data, err = m.fileOps.ReadFile(filepath.Join(m.savePath, "disks.json"))
	if err == nil {
		err = json.Unmarshal(data, &m.DiskIds)
		if err != nil {
			return err
		}
	} else if errors.Is(err, fs.ErrNotExist) {
		m.DiskIds = []model.DiskIdPath{}
	} else {
		return err
	}

	m.syncDisksAndIds()

	return nil
}

func (m *Mgr) saveSettings() error {
	data, err := json.Marshal(m.DiskIds)
	if err != nil {
		return err
	}

	err = m.fileOps.WriteFile(filepath.Join(m.savePath, "disks.json"), data)
	if err != nil {
		return err
	}

	data, err = json.Marshal(m.nodesAddressMap)
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
		case r := <-m.UiMgrConnectTos:
			m.handleConnectToReq(r)
		case r := <-m.UiMgrDisk:
			m.handleAddDiskReq(r)
		case r := <-m.ConnsMgrStatuses:
			m.handleNetConnectedStatus(r)
		case r := <-m.ConnsMgrReceives:
			m.handleReceives(r)
		case r := <-m.DiskMgrReads:
			m.handleDiskReadResult(r)
		case r := <-m.DiskMgrWrites:
			m.handleDiskWriteResult(r)
		case r := <-m.WebdavMgrGets:
			m.handleWebdavGets(r)
		case r := <-m.WebdavMgrPuts:
			m.handleWebdavWriteRequest(r)
		case r := <-m.WebdavMgrBroadcast:
			m.handleWebdavMgrBroadcast(r)
		}
	}
}

func (m *Mgr) handleConnectToReq(i model.UiMgrConnectTo) {
	if _, ok := m.connAddress.Get2(i.Address); !ok {
		chanutil.Send(
			m.ctx,
			m.MgrConnsConnectTos,
			model.MgrConnsConnectTo{Address: string(i.Address)},
			"mgr: handleConnectToReq",
		)
	}
}

func (m *Mgr) syncDisksAndIds() {
	for _, diskIdPath := range m.DiskIds {
		if diskIdPath.Node == m.NodeId {
			m.createDiskChannels(diskIdPath.Id)
			m.createLocalDisk(diskIdPath.Id, diskIdPath.Path)
			m.markLocalDiskAvailable(diskIdPath.Id, diskIdPath.Path, 1)
		} else {
			m.markRemoteDiskUnknown(diskIdPath.Id, diskIdPath.Node, diskIdPath.Path, 1)
		}
	}
}

func (m *Mgr) handleAddDiskReq(i model.AddDiskReq) {
	if i.Node() == m.NodeId {
		id := model.DiskId(uuid.New().String())
		m.DiskIds = append(m.DiskIds, model.DiskIdPath{Id: id, Path: i.Path(), Node: m.NodeId})
		m.syncDisksAndIds()
		err := m.saveSettings()
		if err != nil {
			panic("error saving disk settings")
		}

		for node := range m.nodesAddressMap {
			if conn, exist := m.nodeConnMap.Get1(node); exist {
				iam := model.NewIam(m.NodeId, m.DiskIds, m.nodeAddress, m.freeBytes)
				mcs := model.MgrConnsSend{
					ConnId:  conn,
					Payload: &iam,
				}
				chanutil.Send(m.ctx, m.MgrConnsSends, mcs, "mgr: handleDiskReq: added disk")
			}
		}
	} else {
		if conn, exists := m.nodeConnMap.Get1(i.Node()); exists {
			chanutil.Send(m.ctx, m.MgrConnsSends, model.MgrConnsSend{ConnId: conn, Payload: &i}, "send add disk to node")
		}
	}
}

func (m *Mgr) syncNodesPayloadToSend() model.SyncNodes {
	result := model.NewSyncNodes()
	for node := range m.nodesAddressMap {
		connId, success := m.nodeConnMap.Get1(node)
		if success {
			if address, ok := m.connAddress.Get1(connId); ok {
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
		m.connAddress.Add(i.ConnId, p.Address())
		status := model.UiConnectionStatus{
			Type:          model.Connected,
			RemoteAddress: p.Address(),
			Id:            p.Node(),
		}
		chanutil.Send(m.ctx, m.MgrUiConnectionStatuses, status, "mgr: handleReceives: ui status")
		_ = m.addNodeToCluster(*p, i.ConnId)
		for _, d := range p.Disks() {
			diskStatus := model.UiDiskStatus{
				Localness:     model.Remote,
				Availableness: model.Available,
				Node:          p.Node(),
				Id:            d.Id,
				Path:          d.Path,
			}
			chanutil.Send(m.ctx, m.MgrUiDiskStatuses, diskStatus, "mgr: handleReceives: ui disk status")
		}
		syncNodes := m.syncNodesPayloadToSend()
		for n := range m.nodesAddressMap {
			connId, ok := m.nodeConnMap.Get1(n)
			if ok {
				mcs := model.MgrConnsSend{
					ConnId:  connId,
					Payload: &syncNodes,
				}
				chanutil.Send(m.ctx, m.MgrConnsSends, mcs, "mgr: handleReceives: sync nodes")
			}
		}
	case *model.SyncNodes:
		remoteNodes := p.GetNodes()
		localNodes := set.NewSetFromMapKeys(m.nodesAddressMap)
		localNodes.Add(m.NodeId)
		missing := remoteNodes.Minus(&localNodes)
		for _, n := range missing.GetValues() {
			address := p.AddressForNode(n)
			mct := model.MgrConnsConnectTo{Address: address}
			chanutil.Send(m.ctx, m.MgrConnsConnectTos, mct, "mgr: handleReceives: connect to")

		}
	case *model.WriteRequest:
		caller, ok := m.nodeConnMap.Get2(i.ConnId)
		if ok {
			ptr := p.Data().Ptr
			disk := ptr.Disk()
			chanutil.Send(m.ctx, m.MgrDiskWrites[disk], *p, "mgr: handleReceives: write request")
		} else {
			payload := model.NewWriteResultErr("connection error", caller, p.ReqId())
			mcs := model.MgrConnsSend{
				ConnId:  i.ConnId,
				Payload: &payload,
			}
			chanutil.Send(m.ctx, m.MgrConnsSends, mcs, "mgr: handleReceives: write result conn error")
		}

	case *model.WriteResult:
		m.handleDiskWriteResult(*p)
	case *model.ReadRequest:
		if len(p.Ptrs()) == 0 {
			log.Error("No pointers to read from")
		} else {
			chanutil.Send(m.ctx, m.MgrDiskReads[p.Ptrs()[0].Disk()], *p, "mgr: handleReceives: read request")
		}
	case *model.ReadResult:
		m.handleDiskReadResult(*p)
	case *model.Broadcast:
		chanutil.Send(m.ctx, m.MgrWebdavBroadcast, *p, "mgr: handleReceives: forward broadcast to webdav")
	case *model.AddDiskReq:
		m.handleAddDiskReq(*p)
	default:
		panic("Received unknown payload")
	}
}

func (m *Mgr) handleDiskWriteResult(r model.WriteResult) {
	if r.Caller() == m.NodeId {
		resolved := m.pendingBlockWrites.resolve(r.Ptr(), r.ReqId())
		var err error = nil
		if !r.Ok() {
			err = errors.New(r.Message())
			m.pendingBlockWrites.cancel(r.ReqId())
		}
		switch resolved {
		case done:
			resp := model.PutBlockResp{
				Id:  r.ReqId(),
				Err: err,
			}
			chanutil.Send(m.ctx, m.MgrWebdavPuts, resp, "mgr: handleDiskWriteResult: done")
		}
	} else {
		c, ok := m.nodeConnMap.Get1(r.Caller())
		if ok {
			mcs := model.MgrConnsSend{
				ConnId:  c,
				Payload: &r,
			}
			chanutil.Send(m.ctx, m.MgrConnsSends, mcs, "mgr: handleDiskWriteResult: sending to caller")
		} else {
			panic("not connected")
		}
	}
}

func (m *Mgr) handleDiskReadResult(r model.ReadResult) {
	if r.Ok() {
		if r.Caller() == m.NodeId {
			br := model.GetBlockResp{
				Id: r.ReqId(),
				Block: model.Block{
					Id:   r.BlockId(),
					Type: model.Mirrored,
					Data: r.Data().Data,
				},
			}
			chanutil.Send(m.ctx, m.MgrWebdavGets, br, "mgr: handleDiskReadResult: to local webdav")
		} else {
			c, ok := m.nodeConnMap.Get1(r.Caller())
			if ok {
				mcs := model.MgrConnsSend{ConnId: c, Payload: &r}
				chanutil.Send(m.ctx, m.MgrConnsSends, mcs, "mgr: handleDiskReadResult: to remote webdav")
			} else {
				log.Warn("handleDiskReadResult: not connected")
			}
		}
	} else {
		m.readDiskPtr(r.Ptrs(), r.ReqId(), r.BlockId())
	}
}

func (m *Mgr) addNodeToCluster(iam model.IAm, c model.ConnId) error {
	m.nodesAddressMap[iam.Node()] = iam.Address()
	err := m.saveSettings()
	if err != nil {
		return err
	}
	m.nodeConnMap.Add(iam.Node(), c)
	for _, disk := range iam.Disks() {
		m.mirrorDistributer.SetWeight(iam.Node(), disk.Id, int(iam.FreeBytes()))
	}
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
		address, _ := m.connAddress.Get1(cs.Id)
		id, _ := m.nodeConnMap.Get2(cs.Id)
		m.connAddress.Remove1(cs.Id)
		m.MgrUiConnectionStatuses <- model.UiConnectionStatus{
			Type:          model.NotConnected,
			RemoteAddress: address,
			Msg:           "Disconnected",
			Id:            id,
		}
		// Todo: Need to periodically try to reconnect to unconnected nodes
	}
}

func (m *Mgr) handleWebdavGets(req model.GetBlockReq) {
	ptrs := m.mirrorDistributer.ReadPointersForId(req.BlockId)
	if len(ptrs) == 0 {
		resp := model.GetBlockResp{
			Id:  req.Id(),
			Err: errors.New("not found"),
		}
		chanutil.Send(m.ctx, m.MgrWebdavGets, resp, "mgr: handleWebdavGets: not found")
	} else {
		m.readDiskPtr(ptrs, req.Id(), req.BlockId)
	}
}

func (m *Mgr) readDiskPtr(ptrs []model.DiskPointer, reqId model.GetBlockId, blockId model.BlockId) {
	if len(ptrs) == 0 {
		return
	}
	n := ptrs[0].NodeId()
	disk := ptrs[0].Disk()
	rr := model.NewReadRequest(m.NodeId, ptrs, blockId, reqId)
	if m.NodeId == n {
		chanutil.Send(m.ctx, m.MgrDiskReads[disk], rr, "mgr: readDiskPtr: local")
	} else {
		c, ok := m.nodeConnMap.Get1(n)
		if ok {
			mcs := model.MgrConnsSend{ConnId: c, Payload: &rr}
			chanutil.Send(m.ctx, m.MgrConnsSends, mcs, "mgr: readDiskPtr: remote")
		} else {
			resp := model.GetBlockResp{
				Id:  reqId,
				Err: errors.New("not connected"),
			}
			chanutil.Send(m.ctx, m.MgrWebdavGets, resp, "mgr: readDiskPtr: not connected")
		}
	}
}

func (m *Mgr) handleWebdavWriteRequest(w model.PutBlockReq) {
	switch w.Block.Type {
	case model.Mirrored:
		m.handleMirroredWriteRequest(w)
	case model.XORed:
		panic("unknown block type")
	default:
		panic("unknown block type")
	}
}
func (m *Mgr) handleWebdavMgrBroadcast(b model.Broadcast) {
	for node := range m.nodesAddressMap {
		if connId, exists := m.nodeConnMap.Get1(node); exists {
			chanutil.Send(m.ctx, m.MgrConnsSends, model.MgrConnsSend{ConnId: connId, Payload: &b}, "Broadcasting")
		} else {
			log.Warn("Unable to broadcast to disconnected node")
		}
	}
}

func (m *Mgr) handleMirroredWriteRequest(b model.PutBlockReq) {
	ptrs := m.mirrorDistributer.WritePointersForId(b.Block.Id)
	for _, ptr := range ptrs {
		m.pendingBlockWrites.add(b.Id(), ptr)
		data := model.RawData{
			Data: b.Block.Data,
			Ptr:  ptr,
		}
		writeRequest := model.NewWriteRequest(m.NodeId, data, b.Id())
		if ptr.NodeId() == m.NodeId {
			chanutil.Send(m.ctx, m.MgrDiskWrites[ptr.Disk()], writeRequest, "mgr: handleMirroredWriteRequest: local")
		} else {
			c, ok := m.nodeConnMap.Get1(ptr.NodeId())
			if ok {
				mcs := model.MgrConnsSend{ConnId: c, Payload: &writeRequest}
				chanutil.Send(m.ctx, m.MgrConnsSends, mcs, "mgr: handleMirroredWriteRequest: remote")
			} else {
				m.pendingBlockWrites.cancel(b.Id())
				bir := model.PutBlockResp{Id: b.Id(), Err: errors.New("not connected")}
				chanutil.Send(m.ctx, m.MgrWebdavPuts, bir, "mgr: handleMirroredWriteRequest: not connected")
				return
			}
		}
	}
}
