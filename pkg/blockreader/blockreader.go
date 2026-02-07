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
	"encoding/gob"
	"errors"
	"tealfs/pkg/disk/dist"
	"tealfs/pkg/model"

	log "github.com/sirupsen/logrus"
)

func init() {
	gob.Register(&GetFromDiskReq{})
	gob.Register(&GetFromDiskResp{})
}

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
	return model.GetFromDiskReq
}

type GetFromDiskResp struct {
	Caller model.NodeId
	Dest   Dest
	Resp   model.GetBlockResp
}

func (s *GetFromDiskResp) Type() model.PayloadType {
	return model.GetFromDiskResp
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
		return
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
		bs.LocalDest <- *getFromDisk
	} else {
		bs.RemoteDest <- *getFromDisk
	}
}

func (bs *BlockReader) handleGetResp(requestState map[model.GetBlockId]state, resp GetFromDiskResp) {
	// If we get a response that we don't have record of there isn't much we can do
	if _, ok := requestState[resp.Resp.Id]; ok {
		s := requestState[resp.Resp.Id]
		if resp.Resp.Err == nil {
			// We got the data so we can send it back to the filesystem
			bs.Resp <- resp.Resp
			delete(requestState, resp.Resp.Id)
		} else if len(s.dests) == 0 {
			// If there are no more disks that may have the data we return an error
			delete(requestState, resp.Resp.Id)
			bs.Resp <- model.GetBlockResp{
				Id:    resp.Resp.Id,
				Block: resp.Resp.Block,
				Err:   errors.New("cannot fetch block"),
			}
		} else {
			// If a response is an error then we want to try the next potential dest
			nextDest := s.dests[0]
			s.dests = s.dests[1:]
			requestState[resp.Resp.Id] = s
			bs.sendToLocalOrRemote(&GetFromDiskReq{
				Caller: bs.NodeId,
				Dest:   nextDest,
				Req:    s.req,
			})
		}
	} else {
		log.Warn("Unknown get response")
	}
}
