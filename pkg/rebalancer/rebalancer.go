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
)

type Rebalancer struct {
	InReq         <-chan ExistsReq
	InResp        <-chan ExistsResp
	OutLocal      chan<- ExistsReq
	OutRemote     chan<- model.MgrConnsSend
	OutStoreItCmd chan<- StoreItCmd

	Distributer     *dist.MirrorDistributer
	NodeId          model.NodeId
	Mapper          model.NodeConnectionMapper
	requests        map[ExistsId]set.Set[ExistsReq]
}

func (e *Rebalancer) Start(ctx context.Context) {
	e.requests = make(map[ExistsId]set.Set[ExistsReq])
	for {
		select {
		case <-ctx.Done():
			return
		case req := <-e.InReq:
			e.send(req)
		case resp := <-e.InResp:
			e.handleResp(resp)
		}
	}
}

func (e *Rebalancer) handleResp(resp ExistsResp) {
	if resp.Exists {
		e.unverifiedCount[resp.Req.ExistsId]--
	}
}

func (e *Rebalancer) send(req ExistsReq) {
	writeNodes := e.Distributer.WritePointersForId(req.BlockId)
	e.sentMap[req.ExistsId] = len(writeNodes)
	for _, dest := range writeNodes {
		req.DestNodeId = dest.NodeId
		req.DiskId = dest.Disk
		if req.DestNodeId == e.NodeId {
			e.OutLocal <- req
		} else {
			e.sendRemote(req)
		}
	}
}

func (e *Rebalancer) sendRemote(req ExistsReq) {
	connId, ok := e.Mapper.ConnForNode(req.DestNodeId)
	if ok {
		e.OutRemote <- model.MgrConnsSend{
			Payload: &req,
			ConnId:  connId,
		}
	}
}
