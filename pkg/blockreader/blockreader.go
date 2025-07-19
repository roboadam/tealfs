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

package blockreader

import (
	"context"
	"tealfs/pkg/disk/dist"
	"tealfs/pkg/model"
	"tealfs/pkg/set"
)

type BlockReader struct {
	// Request phase
	Req        <-chan model.GetBlockReq
	RemoteDest chan<- GetFromDiskReq
	LocalDest  chan<- GetFromDiskReq

	// Response phase
	InResp <-chan GetFromDiskResp
	Resp   chan<- model.GetBlockResp

	Distributer *dist.MirrorDistributer
	NodeId      model.NodeId
}

type Dest struct {
	NodeId model.NodeId
	DiskId model.DiskId
}

type GetFromDiskReq struct {
	Caller model.NodeId
	Dest   Dest
	Req    model.GetBlockReq
}

func (s *GetFromDiskReq) Type() model.PayloadType {
	return model.SaveToDiskReq
}

type GetFromDiskResp struct {
	Caller model.NodeId
	Dest   Dest
	Resp   model.GetBlockResp
}

func (s *GetFromDiskResp) Type() model.PayloadType {
	return model.SaveToDiskResp
}

func (bs *BlockReader) Start(ctx context.Context) {
	requestState := make(map[model.GetBlockId]set.Set[model.DiskId])
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

func (bs *BlockReader) handlePutReq(req model.GetBlockReq, requestState map[model.GetBlockId]set.Set[model.DiskId]) {
	// Find all disk destinations for the block
	dests := bs.destsFor(req)
	requestState[req.Id] = set.NewSet[model.DiskId]()

	// For each disk
	for _, dest := range dests {
		// Save each request so we know when we've received all responses
		state := requestState[req.Id]
		state.Add(dest.DiskId)

		getFromDisk := GetFromDiskReq{
			Caller: bs.NodeId,
			Dest:   dest,
			Req: model.GetBlockReq{
				Id:      req.Id,
				BlockId: req.BlockId,
			},
		}

		// If the destination is this node send to the local disk, otherwise send to remote node
		if dest.NodeId == bs.NodeId {
			bs.LocalDest <- getFromDisk
		} else {
			bs.RemoteDest <- getFromDisk
		}
	}
}

func (bs *BlockReader) handleSaveResp(requestState map[model.GetBlockId]set.Set[model.DiskId], resp GetFromDiskResp) {
	// If we get a save response that we don't have record of one of the other destinations must have already failed
	// so we can safely ignore any other responses
	if _, ok := requestState[resp.Resp.Id]; ok {
		if resp.Resp.Err != nil {
			// If a response is an error then we don't have enough redundancy so ignore all following responses and send back
			// an error
			delete(requestState, resp.Resp.Id)
			bs.Resp <- model.GetBlockResp{
				Id:    resp.Resp.Id,
				Block: resp.Resp.Block,
				Err:   resp.Resp.Err,
			}
		} else {
			// If the response is a success remove the record. If all records are removed that means
			// all save requests were successful so we can respond that the save was successful
			state := requestState[resp.Resp.Id]
			state.Remove(resp.Dest.DiskId)
			if state.Len() == 0 {
				delete(requestState, resp.Resp.Id)
				bs.Resp <- model.GetBlockResp{
					Id:    resp.Resp.Id,
					Block: resp.Resp.Block,
				}
			}
		}
	}
}
