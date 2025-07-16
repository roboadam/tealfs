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
	"errors"
	"tealfs/pkg/chanutil"
	"tealfs/pkg/model"
)

func (m *Mgr) handleWebdavGets(req model.GetBlockReq) {
	ptrs := m.MirrorDistributer.ReadPointersForId(req.BlockId)
	if len(ptrs) == 0 {
		m.webdavGetResponseError("not found", req.Id())
	} else {
		m.handleWebdavGetsWithPtrs(ptrs, req.Id(), req.BlockId)
	}
}

func (m *Mgr) webdavGetResponseError(msg string, id model.GetBlockId) {
	resp := model.GetBlockResp{Id: id, Err: errors.New(msg)}
	chanutil.Send(m.ctx, m.MgrWebdavGets, resp, "mgr: handleWebdavGets: "+msg)
}

func (m *Mgr) handleWebdavGetsWithPtrs(ptrs []model.DiskPointer, reqId model.GetBlockId, blockId model.BlockId) {
	if len(ptrs) == 0 {
		m.webdavGetResponseError("exhausted all ptrs", reqId)
		return
	}
	n := ptrs[0].NodeId
	disk := ptrs[0].Disk
	rr := model.ReadRequest{Caller: m.NodeId, Ptrs: ptrs, BlockId: blockId, ReqId: reqId}
	if m.NodeId == n {
		chanutil.Send(m.ctx, m.MgrDiskReads[disk], rr, "mgr: readDiskPtr: local")
	} else {
		c, ok := m.nodeConnMapper.ConnForNode(n)
		if ok {
			mcs := model.MgrConnsSend{ConnId: c, Payload: &rr}
			chanutil.Send(m.ctx, m.MgrConnsSends, mcs, "mgr: readDiskPtr: remote")
		} else {
			m.webdavGetResponseError("not connected", reqId)
		}
	}
}
