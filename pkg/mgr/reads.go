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
		c, ok := m.nodeConnMapper.ConnForNode(n)
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
