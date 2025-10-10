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

	"github.com/google/uuid"
	log "github.com/sirupsen/logrus"
)

type Rebalancer struct {
	InStart       <-chan BalanceReqId
	InResp        <-chan ExistsResp
	OutExistsReq  chan<- ExistsReq
	OutSafeDelete chan<- SafeDelete

	OnFilesystemIds      *set.Map[BalanceReqId, FilesystemBlockIdList]
	NodeId               model.NodeId
	rebalancerMessageMgr RebalancerMessageMgr
	OutStoreItReq        chan<- StoreItReq
	Distributer          *dist.MirrorDistributer
}

func (e *Rebalancer) Start(ctx context.Context) {
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
	if resp.Ok {
		e.rebalancerMessageMgr.removeExistsReq(resp.Req)
		if !e.rebalancerMessageMgr.exists(resp.Req.BalanceReqId, resp.Req.DestBlockId) {
			e.OutSafeDelete <- SafeDelete{
				BalanceReqId: resp.Req.BalanceReqId,
				BlockId:      resp.Req.DestBlockId,
			}
		}
	} else {
		s := StoreItReq{
			Caller:       e.NodeId,
			BalanceReqId: resp.Req.BalanceReqId,
			StoreItId:    StoreItId(uuid.NewString()),
			DestNodeId:   resp.Req.DestNodeId,
			DestDiskId:   resp.Req.DestDiskId,
			DestBlockId:  resp.Req.DestBlockId,
		}
		e.OutStoreItReq <- s
	}

}

func (e *Rebalancer) sendAllExistsReq(balanceReqId BalanceReqId) {
	list, ok := e.OnFilesystemIds.Get(balanceReqId)
	if ok {
		for _, blockId := range list.BlockIds.GetValues() {
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
		e.rebalancerMessageMgr.addExistsReq(req)
		reqs.Add(req)
		e.OutExistsReq <- req
	}
	for _, dest := range writeAndEmpty.Empty {
		d := Dest{
			NodeId: dest.NodeId,
			DiskId: dest.Disk,
		}
		e.rebalancerMessageMgr.addToDelete(balanceReqId, blockId, d.NodeId, d.DiskId)
	}
}
