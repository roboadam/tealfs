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

package blocksaver

import (
	"context"
	"errors"
	"tealfs/pkg/disk/dist"
	"tealfs/pkg/model"
	"testing"

	"github.com/google/uuid"
)

func TestBlockSaver(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	req := make(chan model.PutBlockReq)
	remoteDest := make(chan SaveToDiskReq, 1)
	localDest := make(chan SaveToDiskReq, 1)
	inResp := make(chan SaveToDiskResp)
	resp := make(chan model.PutBlockResp)

	localNodeId := model.NewNodeId()
	localDiskId := model.DiskId(uuid.NewString())
	remoteNodeId := model.NewNodeId()
	remoteDiskId := model.DiskId(uuid.NewString())

	distributer := dist.NewMirrorDistributer(localNodeId)
	distributer.SetWeight(localNodeId, localDiskId, 1)
	distributer.SetWeight(remoteNodeId, remoteDiskId, 1)

	putBlockReq := model.NewPutBlockReq(model.Block{
		Id:   model.NewBlockId(),
		Data: []byte{1, 2, 3, 4, 5},
	})

	bs := BlockSaver{
		Req:         req,
		RemoteDest:  remoteDest,
		LocalDest:   localDest,
		InResp:      inResp,
		Resp:        resp,
		Distributer: &distributer,
		NodeId:      localNodeId,
	}

	go bs.Start(ctx)

	req <- putBlockReq

	localReq := <-localDest
	if localReq.Req.Id != putBlockReq.Id || localReq.Caller != localNodeId {
		t.Error("unexpected req id 1")
		return
	}

	remoteReq := <-remoteDest
	if remoteReq.Req.Id != putBlockReq.Id || remoteReq.Caller != localNodeId {
		t.Error("unexpected req id 2")
		return
	}

	inResp <- SaveToDiskResp{
		Caller: localReq.Caller,
		Dest: Dest{
			NodeId: localNodeId,
			DiskId: localDiskId,
		},
		Resp: model.PutBlockResp{
			Id:  localReq.Req.Id,
			Err: nil,
		},
	}

	select {
	case msg := <-resp:
		t.Errorf("expected no messages, but got: %v", msg)
	default:
	}

	inResp <- SaveToDiskResp{
		Caller: remoteReq.Caller,
		Dest: Dest{
			NodeId: remoteNodeId,
			DiskId: remoteDiskId,
		},
		Resp: model.PutBlockResp{
			Id:  remoteReq.Req.Id,
			Err: nil,
		},
	}

	msg := <-resp

	if msg.Id != putBlockReq.Id {
		t.Error("didn't get final response")
	}

	req <- putBlockReq
	localReq = <-localDest
	remoteReq = <-remoteDest

	inResp <- SaveToDiskResp{
		Caller: localReq.Caller,
		Dest: Dest{
			NodeId: localNodeId,
			DiskId: localDiskId,
		},
		Resp: model.PutBlockResp{
			Id:  localReq.Req.Id,
			Err: errors.New("some error putting the first one"),
		},
	}

	msg = <-resp

	if msg.Id != putBlockReq.Id || msg.Err == nil {
		t.Error("didn't get error response")
	}
}
