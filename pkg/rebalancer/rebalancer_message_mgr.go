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

package rebalancer

import (
	"tealfs/pkg/model"
	"tealfs/pkg/set"
)

type RebalancerMessageMgr struct {
	pendingExists map[BalanceReqId]map[model.BlockId]set.Set[ExistsReq]
	toDelete      map[BalanceReqId]map[model.BlockId]set.Set[Dest]
}

func (r *RebalancerMessageMgr) init() {
	if r.pendingExists == nil {
		r.pendingExists = make(map[BalanceReqId]map[model.BlockId]set.Set[ExistsReq])
		r.toDelete = make(map[BalanceReqId]map[model.BlockId]set.Set[Dest])
	}
}

func (r *RebalancerMessageMgr) initBalanceReq(balanceReqId BalanceReqId) {
	r.init()
	if _, ok := r.pendingExists[balanceReqId]; ok {
		return
	}
	r.pendingExists[balanceReqId] = make(map[model.BlockId]set.Set[ExistsReq])
	r.toDelete[balanceReqId] = make(map[model.BlockId]set.Set[Dest])
}

func (r *RebalancerMessageMgr) initBlockId(balanceReqId BalanceReqId, blockId model.BlockId) {
	r.initBalanceReq(balanceReqId)
	if _, ok := r.pendingExists[balanceReqId][blockId]; ok {
		return
	}
	r.pendingExists[balanceReqId][blockId] = set.NewSet[ExistsReq]()
	r.toDelete[balanceReqId][blockId] = set.NewSet[Dest]()
}

func (r *RebalancerMessageMgr) addExistsReq(req ExistsReq) {
	r.initBlockId(req.BalanceReqId, req.DestBlockId)
	s := r.pendingExists[req.BalanceReqId][req.DestBlockId]
	s.Add(req)
	r.pendingExists[req.BalanceReqId][req.DestBlockId] = s
}

func (r *RebalancerMessageMgr) removeExistsReq(req ExistsReq) {
	r.initBlockId(req.BalanceReqId, req.DestBlockId)
	s := r.pendingExists[req.BalanceReqId][req.DestBlockId]
	s.Remove(req)
	r.pendingExists[req.BalanceReqId][req.DestBlockId] = s
}

func (r *RebalancerMessageMgr) exists(balanceReqId BalanceReqId, blockId model.BlockId) bool {
	r.initBlockId(balanceReqId, blockId)
	s := r.pendingExists[balanceReqId][blockId]
	return s.Len() == 0
}

func (r *RebalancerMessageMgr) addToDelete(balanceReqId BalanceReqId, blockId model.BlockId, nodeId model.NodeId, diskId model.DiskId) {
	r.initBlockId(balanceReqId, blockId)
	d := Dest{
		NodeId: nodeId,
		DiskId: diskId,
	}
	s := r.toDelete[balanceReqId][blockId]
	s.Add(d)
	r.toDelete[balanceReqId][blockId] = s
}

func (r *RebalancerMessageMgr) removeToDelete(balanceReqId BalanceReqId, blockId model.BlockId, nodeId model.NodeId, diskId model.DiskId) {
	r.initBlockId(balanceReqId, blockId)
	d := Dest{
		NodeId: nodeId,
		DiskId: diskId,
	}
	s := r.toDelete[balanceReqId][blockId]
	s.Remove(d)
	r.toDelete[balanceReqId][blockId] = s
}
