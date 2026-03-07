// Copyright (C) 2026 Adam Hess
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

package blocksaver

import (
	"context"
	"encoding/gob"
	"tealfs/pkg/datalayer"
	"tealfs/pkg/model"
	"tealfs/pkg/set"
)

func init() {
	gob.Register(&SaveToDiskReq{})
	gob.Register(&SaveToDiskResp{})
}

// This process accepts put requests from the local filesystem process only
// Then it figures which disk (remote or local) need to get the data, then it send
// The SaveToDiskReq to the appropriate places
// It then accepts responses from the disks and once all are successful it responds
// to the local filesystem with a success message

type BlockSaver struct {
	// Request phase
	Req        <-chan model.PutBlockReq
	RemoteDest chan<- SaveToDiskReq
	LocalDest  chan<- SaveToDiskReq

	// Response phase
	InResp <-chan SaveToDiskResp
	Resp   chan<- model.PutBlockResp

	NodeId       model.NodeId
	Disks        *set.Set[model.DiskInfo]
	StateHandler *datalayer.StateHandler
}

type Dest struct {
	NodeId model.NodeId
	DiskId model.DiskId
}

type SaveToDiskReq struct {
	Caller model.NodeId
	Dest   Dest
	Req    model.PutBlockReq
}

type SaveToDiskResp struct {
	Caller model.NodeId
	Dest   Dest
	Resp   model.PutBlockResp
}

func (bs *BlockSaver) Start(ctx context.Context) {
	requestState := make(map[model.PutBlockId]model.DiskId)
	for {
		select {
		case req := <-bs.Req:
			bs.handlePutReq(req, requestState)
		case resp := <-bs.InResp:
			bs.handleSaveResp(requestState, resp)
		case <-ctx.Done():
			return
		}
	}
}

func (bs *BlockSaver) handlePutReq(req model.PutBlockReq, requestState map[model.PutBlockId]model.DiskId) {
	// Save each request so we know when we've received all responses
	dest := bs.dest()
	requestState[req.Id] = dest.DiskId

	saveToDisk := SaveToDiskReq{Dest: dest, Req: req, Caller: bs.NodeId}

	// If the destination is this node send to the local disk, otherwise send to remote node
	if dest.NodeId == bs.NodeId {
		bs.LocalDest <- saveToDisk
	} else {
		bs.RemoteDest <- saveToDisk
	}
}

func (bs *BlockSaver) dest() Dest {
	disks := bs.Disks.GetValues()
	for _, d := range disks {
		if d.NodeId == bs.NodeId {
			return Dest{
				NodeId: d.NodeId,
				DiskId: d.DiskId,
			}
		}
	}
	return Dest{
		NodeId: disks[0].NodeId,
		DiskId: disks[0].DiskId,
	}
}

func (bs *BlockSaver) handleSaveResp(requestState map[model.PutBlockId]model.DiskId, resp SaveToDiskResp) {
	if _, ok := requestState[resp.Resp.Id]; ok {
		delete(requestState, resp.Resp.Id)
		if resp.Resp.Err != nil {
			delete(requestState, resp.Resp.Id)
			bs.Resp <- model.PutBlockResp{
				Id:  resp.Resp.Id,
				Err: resp.Resp.Err,
			}
		} else {
			bs.Resp <- model.PutBlockResp{
				Id: resp.Resp.Id,
			}
			bs.StateHandler.Saved(resp.Resp.)
		}
	}
}
