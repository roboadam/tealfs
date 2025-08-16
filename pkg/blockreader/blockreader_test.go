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
	"testing"

	"github.com/google/uuid"
)

func TestNoDisks(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	req := make(chan model.GetBlockReq)
	remoteDest := make(chan GetFromDiskReq)
	localDest := make(chan GetFromDiskReq)
	inResp := make(chan GetFromDiskResp)
	resp := make(chan model.GetBlockResp)
	localNodeId := model.NewNodeId()
	distributer := dist.NewMirrorDistributer(localNodeId)

	br := BlockReader{
		Req:         req,
		RemoteDest:  remoteDest,
		LocalDest:   localDest,
		InResp:      inResp,
		Resp:        resp,
		Distributer: &distributer,
		NodeId:      localNodeId,
	}

	go br.Start(ctx)

	blockId := model.NewBlockId()
	noDiskReq := model.NewGetBlockReq(blockId)
	req <- noDiskReq
	noDiskResp := <-resp
	if noDiskResp.Id != noDiskReq.Id || noDiskResp.Err == nil {
		t.Error("result should be a block id with an error")
		return
	}
}

func TestLocalSuccess(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	req := make(chan model.GetBlockReq)
	remoteDest := make(chan GetFromDiskReq)
	localDest := make(chan GetFromDiskReq)
	inResp := make(chan GetFromDiskResp)
	resp := make(chan model.GetBlockResp)
	localNodeId := model.NewNodeId()
	localDiskId := model.DiskId(uuid.NewString())
	remoteNodeId := model.NewNodeId()
	remoteDiskId := model.DiskId(uuid.NewString())
	distributer := dist.NewMirrorDistributer(localNodeId)

	br := BlockReader{
		Req:         req,
		RemoteDest:  remoteDest,
		LocalDest:   localDest,
		InResp:      inResp,
		Resp:        resp,
		Distributer: &distributer,
		NodeId:      localNodeId,
	}

	go br.Start(ctx)

	distributer.SetWeight(localNodeId, localDiskId, 1)
	distributer.SetWeight(remoteNodeId, remoteDiskId, 1)

	blockId := model.NewBlockId()
	localSuccessReq := model.NewGetBlockReq(blockId)
	req <- localSuccessReq

	select {
	case <-remoteDest:
		t.Error("Should request local first")
		return
	default:
	}

	localReq := <-localDest
	if localReq.Req.Id != localSuccessReq.Id {
		t.Error("Invalid block id")
		return
	}

	inResp <- GetFromDiskResp{
		Caller: localReq.Caller,
		Dest:   localReq.Dest,
		Resp: model.GetBlockResp{
			Id: localReq.Req.Id,
			Block: model.Block{
				Id:   blockId,
				Data: []byte{1, 2, 3, 4},
			},
		},
	}

	localSuccessResp := <-resp
	if localSuccessReq.Id != localSuccessResp.Id {
		t.Error("Invalid request id")
		return
	}
}

func TestLocalErrorRemoteSuccess(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	req := make(chan model.GetBlockReq)
	remoteDest := make(chan GetFromDiskReq)
	localDest := make(chan GetFromDiskReq)
	inResp := make(chan GetFromDiskResp)
	resp := make(chan model.GetBlockResp)
	localNodeId := model.NewNodeId()
	localDiskId := model.DiskId(uuid.NewString())
	remoteNodeId := model.NewNodeId()
	remoteDiskId := model.DiskId(uuid.NewString())
	distributer := dist.NewMirrorDistributer(localNodeId)

	br := BlockReader{
		Req:         req,
		RemoteDest:  remoteDest,
		LocalDest:   localDest,
		InResp:      inResp,
		Resp:        resp,
		Distributer: &distributer,
		NodeId:      localNodeId,
	}

	go br.Start(ctx)

	distributer.SetWeight(localNodeId, localDiskId, 1)
	distributer.SetWeight(remoteNodeId, remoteDiskId, 1)

	blockId := model.NewBlockId()
	remoteSuccessReq := model.NewGetBlockReq(blockId)
	req <- remoteSuccessReq

	localReq := <-localDest

	inResp <- GetFromDiskResp{
		Caller: localReq.Caller,
		Dest:   localReq.Dest,
		Resp: model.GetBlockResp{
			Id:  localReq.Req.Id,
			Err: errors.New("some error on the local disk"),
		},
	}

	remoteReq := <-remoteDest
	if remoteReq.Req.Id != remoteSuccessReq.Id && remoteReq.Dest.NodeId == remoteNodeId {
		t.Error("invalid id")
		return
	}

	inResp <- GetFromDiskResp{
		Caller: localNodeId,
		Dest:   remoteReq.Dest,
		Resp: model.GetBlockResp{
			Id: remoteReq.Req.Id,
			Block: model.Block{
				Id:   blockId,
				Data: []byte{1, 2, 3, 4, 5},
			},
		},
	}

	remoteSuccessResp := <-resp
	if remoteSuccessReq.Id != remoteSuccessResp.Id || remoteSuccessResp.Err != nil {
		t.Error("Invalid request id or error resp")
		return
	}
}

func TestLocalErrorRemoteError(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	req := make(chan model.GetBlockReq)
	remoteDest := make(chan GetFromDiskReq)
	localDest := make(chan GetFromDiskReq)
	inResp := make(chan GetFromDiskResp)
	resp := make(chan model.GetBlockResp)
	localNodeId := model.NewNodeId()
	localDiskId := model.DiskId(uuid.NewString())
	remoteNodeId := model.NewNodeId()
	remoteDiskId := model.DiskId(uuid.NewString())
	distributer := dist.NewMirrorDistributer(localNodeId)

	br := BlockReader{
		Req:         req,
		RemoteDest:  remoteDest,
		LocalDest:   localDest,
		InResp:      inResp,
		Resp:        resp,
		Distributer: &distributer,
		NodeId:      localNodeId,
	}

	go br.Start(ctx)

	distributer.SetWeight(localNodeId, localDiskId, 1)
	distributer.SetWeight(remoteNodeId, remoteDiskId, 1)

	blockId := model.NewBlockId()
	remoteSuccessReq := model.NewGetBlockReq(blockId)
	req <- remoteSuccessReq

	localReq := <-localDest

	inResp <- GetFromDiskResp{
		Caller: localReq.Caller,
		Dest:   localReq.Dest,
		Resp: model.GetBlockResp{
			Id:  localReq.Req.Id,
			Err: errors.New("some error on the local disk"),
		},
	}

	remoteReq := <-remoteDest

	inResp <- GetFromDiskResp{
		Caller: localNodeId,
		Dest:   remoteReq.Dest,
		Resp: model.GetBlockResp{
			Id:  remoteReq.Req.Id,
			Err: errors.New("some error on the remote disk"),
		},
	}

	remoteError := <-resp
	if remoteSuccessReq.Id != remoteError.Id || remoteError.Err == nil {
		t.Error("Invalid request id or not error")
		return
	}
}
