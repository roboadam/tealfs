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
	"errors"
	"tealfs/pkg/model"
	"tealfs/pkg/set"
	"testing"

	"github.com/google/uuid"
)

func TestBlockSaver(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	req := make(chan model.PutBlockReq)
	outPayloads := make(chan model.Payload2, 1)
	inResp := make(chan SaveToDiskResp)
	resp := make(chan model.PutBlockResp)

	localNodeId := model.NewNodeId()
	localDiskId := model.DiskId(uuid.NewString())
	remoteNodeId := model.NewNodeId()
	remoteDiskId := model.DiskId(uuid.NewString())

	putBlockReq := model.NewPutBlockReq(model.Block{
		Id:   model.NewBlockId(),
		Data: []byte{1, 2, 3, 4, 5},
	})

	diskInfoSet := set.NewSet[model.DiskInfo]()
	diskInfoSet.Add(model.DiskInfo{
		DiskId: localDiskId,
		Path:   "localPath",
		NodeId: localNodeId,
	})
	diskInfoSet.Add(model.DiskInfo{
		DiskId: remoteDiskId,
		Path:   "remotePath",
		NodeId: remoteNodeId,
	})

	bs := BlockSaver{
		Req:          req,
		InResp:       inResp,
		Resp:         resp,
		NodeId:       localNodeId,
		DiskInfoList: &diskInfoSet,
		OutPayload:   outPayloads,
	}

	go bs.Start(ctx)

	req <- putBlockReq

	payload := <-outPayloads
	saveReq, ok := payload.(*SaveToDiskReq)
	if !ok {
		t.Fatal("wrong type")
	}

	if saveReq.Req.Id != putBlockReq.Id || saveReq.Caller != localNodeId {
		t.Error("unexpected req id 1")
		return
	}

	inResp <- SaveToDiskResp{
		Caller: saveReq.Caller,
		Dest: Dest{
			NodeId: localNodeId,
			DiskId: localDiskId,
		},
		Resp: model.PutBlockResp{
			Id:  saveReq.Req.Id,
			Err: nil,
		},
	}

	select {
	case msg := <-resp:
		t.Errorf("expected no messages, but got: %v", msg)
	default:
	}

	msg := <-resp

	if msg.Id != putBlockReq.Id {
		t.Error("didn't get final response")
	}

	req <- putBlockReq
	payload = <-outPayloads
	saveReq, ok = payload.(*SaveToDiskReq)
	if !ok {
		t.Fatal("wrong type")
	}

	inResp <- SaveToDiskResp{
		Caller: saveReq.Caller,
		Dest: Dest{
			NodeId: localNodeId,
			DiskId: localDiskId,
		},
		Resp: model.PutBlockResp{
			Id:  saveReq.Req.Id,
			Err: errors.New("some error putting the first one"),
		},
	}

	msg = <-resp

	if msg.Id != putBlockReq.Id || msg.Err == nil {
		t.Error("didn't get error response")
	}
}
