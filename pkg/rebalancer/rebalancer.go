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
	"context"
	"tealfs/pkg/disk/dist"
	"tealfs/pkg/model"
	"tealfs/pkg/set"

	log "github.com/sirupsen/logrus"
)

type Rebalancer struct {
	InStart      <-chan BalanceReqId
	OutExistsReq chan<- ExistsReq

	OnFilesystemIds *set.Map[BalanceReqId, FilesystemBlockIdList]
	pendingExists   map[BalanceReqId]map[model.BlockId]*set.Set[ExistsReq]
	toDelete        map[BalanceReqId]map[model.BlockId]*set.Set[Dest]

	//////////////////////

	InResp        <-chan ExistsResp
	OutRemote     chan<- model.MgrConnsSend
	OutStoreItCmd chan<- StoreItCmd

	Distributer *dist.MirrorDistributer
	NodeId      model.NodeId
}

func (e *Rebalancer) Start(ctx context.Context) {
	e.pendingExists = make(map[BalanceReqId]map[model.BlockId]*set.Set[ExistsReq])
	for {
		select {
		case <-ctx.Done():
			return
		case req := <-e.InStart:
			e.sendAllExistsReq(req)
		case resp := <-e.InResp:
			e.handleResp(resp)
		}
	}
}

func (e *Rebalancer) handleResp(resp ExistsResp) {
}

func (e *Rebalancer) sendAllExistsReq(balanceReqId BalanceReqId) {
	list, ok := e.OnFilesystemIds.Get(balanceReqId)
	if ok {
		e.pendingExists[balanceReqId] = make(map[model.BlockId]*set.Set[ExistsReq])
		e.toDelete[balanceReqId] = make(map[model.BlockId]*set.Set[Dest])
		for _, blockId := range list.BlockIds.GetValues() {
			e.pendingExists[balanceReqId][blockId] = mak
			e.sendExistsReq(blockId, balanceReqId)
		}
	} else {
		log.Warn("key not found")
	}
}

func (e *Rebalancer) sendExistsReq(blockId model.BlockId, balanceReqId BalanceReqId) {
	writeAndEmpty := e.Distributer.WriteAndEmptyPtrs(blockId)
	reqs := set.NewSet[ExistsReq]()
	for _, dest := range writeAndEmpty.Write {
		req := ExistsReq{
			Caller:       e.NodeId,
			BalanceReqId: balanceReqId,
			ExistsId:     ExistsId(blockId),
			DestNodeId:   dest.NodeId,
			DestDiskId:   dest.Disk,
			DestBlockId:  blockId,
		}
		reqs.Add(req)
		e.OutExistsReq <- req
	}
	for _, dest := range writeAndEmpty.Empty {
		d := Dest{
			NodeId: dest.NodeId,
			DiskId: dest.Disk,
		}
		e.toDelete[balanceReqId][blockId].Add(d)
	}
}
