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
	UiMgrConnectTos    chan model.UiMgrConnectTo
	ConnsMgrStatuses   chan model.NetConnectionStatus
	ConnsMgrReceives   chan model.ConnsMgrReceive
	DiskMgrReads       chan model.ReadResult
	DiskMgrWrites      chan model.WriteResult
	WebdavMgrGets      chan model.GetBlockReq
	WebdavMgrPuts      chan model.PutBlockReq
	WebdavMgrBroadcast chan model.Broadcast
	MgrConnsConnectTos chan model.MgrConnsConnectTo
	MgrConnsSends      chan model.MgrConnsSend
	MgrDiskWrites      map[model.DiskId]chan model.WriteRequest
	MgrDiskReads       map[model.DiskId]chan model.ReadRequest
	MgrUiStatuses      chan model.UiConnectionStatus
	MgrWebdavGets      chan model.GetBlockResp
	MgrWebdavPuts      chan model.PutBlockResp
	MgrWebdavBroadcast chan model.Broadcast

	nodesAddressMap    map[model.NodeId]string
	nodeConnMap        set.Bimap[model.NodeId, model.ConnId]
	NodeId             model.NodeId
	connAddress        map[model.ConnId]string
	mirrorDistributer  dist.MirrorDistributer
	blockType          model.BlockType
	nodeAddress        string
	savePath           string
	fileOps            disk.FileOps
	pendingBlockWrites pendingBlockWrites
	freeBytes          uint32
	disks              map[model.DiskId]string
	DiskIds            []model.DiskId
}

func NewWithChanSize(
	chanSize int,
	nodeAddress string,
	globalPath string,
	fileOps disk.FileOps,
	blockType model.BlockType,
	freeBytes uint32,
	disks []string,
) *Mgr {
	nodeId, err := readNodeId(globalPath, fileOps)
	if err != nil {
		panic(err)
	}

	mgr := Mgr{
		UiMgrConnectTos:    make(chan model.UiMgrConnectTo, chanSize),
		ConnsMgrStatuses:   make(chan model.NetConnectionStatus, chanSize),
		ConnsMgrReceives:   make(chan model.ConnsMgrReceive, chanSize),
		DiskMgrWrites:      make(chan model.WriteResult),
		DiskMgrReads:       make(chan model.ReadResult, chanSize),
		WebdavMgrGets:      make(chan model.GetBlockReq, chanSize),
		WebdavMgrPuts:      make(chan model.PutBlockReq, chanSize),
		WebdavMgrBroadcast: make(chan model.Broadcast, chanSize),
		MgrConnsConnectTos: make(chan model.MgrConnsConnectTo, chanSize),
		MgrConnsSends:      make(chan model.MgrConnsSend, chanSize),
		MgrUiStatuses:      make(chan model.UiConnectionStatus, chanSize),
		MgrWebdavGets:      make(chan model.GetBlockResp, chanSize),
		MgrWebdavPuts:      make(chan model.PutBlockResp, chanSize),
		MgrWebdavBroadcast: make(chan model.Broadcast, chanSize),
		nodesAddressMap:    make(map[model.NodeId]string),
		NodeId:             nodeId,
		connAddress:        make(map[model.ConnId]string),
		nodeConnMap:        set.NewBimap[model.NodeId, model.ConnId](),
		mirrorDistributer:  dist.NewMirrorDistributer(),
		blockType:          blockType,
		nodeAddress:        nodeAddress,
		savePath:           globalPath,
		fileOps:            fileOps,
		pendingBlockWrites: newPendingBlockWrites(),
		freeBytes:          freeBytes,
		disks:              make(map[model.DiskId]string),
		DiskIds:            []model.DiskId{},
	}

	err = mgr.loadSettings(disks)
	if err != nil {
		panic("Unable to load settings")
	}
	mgr.MgrDiskWrites = diskWriteChans(mgr.disks)
	mgr.MgrDiskReads = diskReadChans(mgr.disks)

	for disk := range mgr.disks {
		mgr.DiskIds = append(mgr.DiskIds, disk)
		mgr.mirrorDistributer.SetWeight(mgr.NodeId, disk, int(freeBytes))
	}

	return &mgr
}

func diskWriteChans(ids map[model.DiskId]string) map[model.DiskId]chan model.WriteRequest {
	return diskChans[model.WriteRequest](ids)
}

func diskReadChans(ids map[model.DiskId]string) map[model.DiskId]chan model.ReadRequest {
	return diskChans[model.ReadRequest](ids)
}

func diskChans[V any](ids map[model.DiskId]string) map[model.DiskId]chan V {
	result := make(map[model.DiskId]chan V)
	for id := range ids {
		result[id] = make(chan V)
	}
	return result
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

func (m *Mgr) Start(ctx context.Context) error {
	go m.eventLoop(ctx)
	for nodeId, address := range m.nodesAddressMap {
		if nodeId != m.NodeId {
			m.UiMgrConnectTos <- model.UiMgrConnectTo{
				Address: address,
			}
		}
	}
	return nil
}

func (m *Mgr) loadSettings(diskPaths []string) error {
	data, err := m.fileOps.ReadFile(filepath.Join(m.savePath, "disks.json"))
	if errors.Is(err, fs.ErrNotExist) {
		for _, path := range diskPaths {
			id := model.DiskId(uuid.New().String())
			m.disks[id] = path
		}
	} else if err != nil {
		return err
	} else {
		err = json.Unmarshal(data, &m.disks)
		if err != nil {
			return err
		}
	}

	data, err = m.fileOps.ReadFile(filepath.Join(m.savePath, "cluster.json"))
	if err != nil && !errors.Is(err, fs.ErrNotExist) {
		return err
	}
	if len(data) > 0 {
		err = json.Unmarshal(data, &m.nodesAddressMap)
		if err != nil {
			return err
		}
	}
	err = m.saveSettings()
	if err != nil {
		return err
	}

	return nil
}

func (m *Mgr) saveSettings() error {
	data, err := json.Marshal(m.disks)
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

func (m *Mgr) eventLoop(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			return
		case r := <-m.UiMgrConnectTos:
			m.handleConnectToReq(r)
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
	chanutil.Send(
		m.MgrConnsConnectTos,
		model.MgrConnsConnectTo{Address: string(i.Address)},
		"mgr: handleConnectToReq",
	)
}

func (m *Mgr) syncNodesPayloadToSend() model.SyncNodes {
	result := model.NewSyncNodes()
	for node := range m.nodesAddressMap {
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
		m.connAddress[i.ConnId] = p.Address()
		status := model.UiConnectionStatus{
			Type:          model.Connected,
			RemoteAddress: p.Address(),
			Id:            p.Node(),
		}
		chanutil.Send(m.MgrUiStatuses, status, "mgr: handleReceives: ui status")
		_ = m.addNodeToCluster(*p, i.ConnId)
		syncNodes := m.syncNodesPayloadToSend()
		for n := range m.nodesAddressMap {
			connId, ok := m.nodeConnMap.Get1(n)
			if ok {
				mcs := model.MgrConnsSend{
					ConnId:  connId,
					Payload: &syncNodes,
				}
				chanutil.Send(m.MgrConnsSends, mcs, "mgr: handleReceives: sync nodes")
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
			chanutil.Send(m.MgrConnsConnectTos, mct, "mgr: handleReceives: connect to")

		}
	case *model.WriteRequest:
		caller, ok := m.nodeConnMap.Get2(i.ConnId)
		if ok {
			ptr := p.Data().Ptr
			disk := ptr.Disk()
			chanutil.Send(m.MgrDiskWrites[disk], *p, "mgr: handleReceives: write request")
		} else {
			payload := model.NewWriteResultErr("connection error", caller, p.ReqId())
			mcs := model.MgrConnsSend{
				ConnId:  i.ConnId,
				Payload: &payload,
			}
			chanutil.Send(m.MgrConnsSends, mcs, "mgr: handleReceives: write result conn error")
		}

	case *model.WriteResult:
		m.handleDiskWriteResult(*p)
	case *model.ReadRequest:
		if len(p.Ptrs()) == 0 {
			log.Error("No pointers to read from")
		} else {
			chanutil.Send(m.MgrDiskReads[p.Ptrs()[0].Disk()], *p, "mgr: handleReceives: read request")
		}
	case *model.ReadResult:
		m.handleDiskReadResult(*p)
	case *model.Broadcast:
		chanutil.Send(m.MgrWebdavBroadcast, *p, "mgr: handleReceives: forward broadcast to webdav")
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
			chanutil.Send(m.MgrWebdavPuts, resp, "mgr: handleDiskWriteResult: done")
		}
	} else {
		c, ok := m.nodeConnMap.Get1(r.Caller())
		if ok {
			mcs := model.MgrConnsSend{
				ConnId:  c,
				Payload: &r,
			}
			chanutil.Send(m.MgrConnsSends, mcs, "mgr: handleDiskWriteResult: sending to caller")
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
			chanutil.Send(m.MgrWebdavGets, br, "mgr: handleDiskReadResult: to local webdav")
		} else {
			c, ok := m.nodeConnMap.Get1(r.Caller())
			if ok {
				mcs := model.MgrConnsSend{ConnId: c, Payload: &r}
				chanutil.Send(m.MgrConnsSends, mcs, "mgr: handleDiskReadResult: to remote webdav")
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
		m.mirrorDistributer.SetWeight(iam.Node(), disk, int(iam.FreeBytes()))
	}
	return nil
}

func (m *Mgr) Disks() map[model.DiskId]string {
	return m.disks
}

func (m *Mgr) handleNetConnectedStatus(cs model.NetConnectionStatus) {
	switch cs.Type {
	case model.Connected:
		iam := model.NewIam(m.NodeId, m.DiskIds, m.nodeAddress, m.freeBytes)
		mcs := model.MgrConnsSend{
			ConnId:  cs.Id,
			Payload: &iam,
		}
		chanutil.Send(m.MgrConnsSends, mcs, "mgr: handleNetConnectedStatus: connected")
	case model.NotConnected:
		address := m.connAddress[cs.Id]
		delete(m.connAddress, cs.Id)
		// Todo: need a mechanism to back off
		ct := model.MgrConnsConnectTo{Address: address}
		chanutil.Send(m.MgrConnsConnectTos, ct, "mgr: handleNetConnectedStatus: not connected")
		// Todo: reflect this in the ui
	}
}

func (m *Mgr) handleWebdavGets(req model.GetBlockReq) {
	ptrs := m.mirrorDistributer.ReadPointersForId(req.BlockId)
	if len(ptrs) == 0 {
		resp := model.GetBlockResp{
			Id:  req.Id(),
			Err: errors.New("not found"),
		}
		chanutil.Send(m.MgrWebdavGets, resp, "mgr: handleWebdavGets: not found")
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
		chanutil.Send(m.MgrDiskReads[disk], rr, "mgr: readDiskPtr: local")
	} else {
		c, ok := m.nodeConnMap.Get1(n)
		if ok {
			mcs := model.MgrConnsSend{ConnId: c, Payload: &rr}
			chanutil.Send(m.MgrConnsSends, mcs, "mgr: readDiskPtr: remote")
		} else {
			resp := model.GetBlockResp{
				Id:  reqId,
				Err: errors.New("not connected"),
			}
			chanutil.Send(m.MgrWebdavGets, resp, "mgr: readDiskPtr: not connected")
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
		if connid, exists := m.nodeConnMap.Get1(node); exists {
			chanutil.Send(m.MgrConnsSends, model.MgrConnsSend{ConnId: connid, Payload: &b}, "Broadcasting")
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
			chanutil.Send(m.MgrDiskWrites[ptr.Disk()], writeRequest, "mgr: handleMirroredWriteRequest: local "+m.disks[ptr.Disk()]+" "+string(ptr.Disk()))
		} else {
			c, ok := m.nodeConnMap.Get1(ptr.NodeId())
			if ok {
				mcs := model.MgrConnsSend{ConnId: c, Payload: &writeRequest}
				chanutil.Send(m.MgrConnsSends, mcs, "mgr: handleMirroredWriteRequest: remote")
			} else {
				m.pendingBlockWrites.cancel(b.Id())
				bir := model.PutBlockResp{Id: b.Id(), Err: errors.New("not connected")}
				chanutil.Send(m.MgrWebdavPuts, bir, "mgr: handleMirroredWriteRequest: not connected")
				return
			}
		}
	}
}
