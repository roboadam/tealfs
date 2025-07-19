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
	"errors"
	"tealfs/pkg/disk/dist"
	"tealfs/pkg/model"
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
	requestState := make(map[model.GetBlockId]state)
	for {
		select {
		case req := <-bs.Req:
			bs.handleGetReq(req, requestState)
		case resp := <-bs.InResp:
			bs.handleGetResp(requestState, resp)
		case <-ctx.Done():
			return
		}
	}
}

type state struct {
	req   model.GetBlockReq
	dests []Dest
}

func (bs *BlockReader) handleGetReq(req model.GetBlockReq, requestState map[model.GetBlockId]state) {
	// Find all potential disk destinations for the block
	dests := bs.destsFor(req)

	// If there are no disks to write to then reply with an error
	if len(dests) == 0 {
		bs.Resp <- model.GetBlockResp{
			Id:  req.Id,
			Err: errors.New("no dests"),
		}
	}

	firstDest := dests[0]
	dests = dests[1:]

	// Request state hold a list of dests we haven't tried yet
	requestState[req.Id] = state{
		req:   req,
		dests: dests,
	}

	getFromDisk := GetFromDiskReq{
		Caller: bs.NodeId,
		Dest:   firstDest,
		Req: model.GetBlockReq{
			Id:      req.Id,
			BlockId: req.BlockId,
		},
	}

	bs.sendToLocalOrRemote(&getFromDisk)
}

func (bs *BlockReader) sendToLocalOrRemote(getFromDisk *GetFromDiskReq) {
	if getFromDisk.Dest.NodeId == bs.NodeId {
		bs.LocalDest <- getFromDisk
	} else {
		bs.RemoteDest <- getFromDisk
	}
}

func (bs *BlockReader) handleGetResp(requestState map[model.GetBlockId]state, resp GetFromDiskResp) {
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
			reqId := resp.Resp.Id
			if len(requestState[reqId].dests) > 0 {
				dest := requestState[reqId].dests[0]
				getFromDisk := GetFromDiskReq{
					Caller: bs.NodeId,
					Dest:   dest,
					Req:    requestState[reqId].req,
				}
				bs.sendToLocalOrRemote(&getFromDisk)
			} else {

			}
		}
	}
}
