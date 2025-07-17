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
	"tealfs/pkg/disk"
	"tealfs/pkg/model"
	"testing"

	"github.com/google/uuid"
)

func TestLocalBlockSaveResponse(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	nodeId := model.NewNodeId()
	remoteNodeId := model.NewNodeId()

	disk1 := mockDisk(nodeId, ctx)
	disk2 := mockDisk(nodeId, ctx)

	resp := make(chan SaveToDiskResp)
	sends := make(chan model.MgrConnsSend)
	addedDisks := make(chan *disk.Disk)

	lbsr := LocalBlockSaveResponses{
		InDisks:             addedDisks,
		LocalWriteResponses: resp,
		Sends:               sends,
		NodeConnMap:         model.NewNodeConnectionMapper(),
		NodeId:              nodeId,
	}

	go lbsr.Start(ctx)

	addedDisks <- disk1
	addedDisks <- disk2
	lbsr.NodeConnMap.SetAll(1, "some-address:123", remoteNodeId)

	putBlockId := model.PutBlockId(uuid.NewString())
	disk1.OutWrites <- model.NewWriteResultOk(
		model.DiskPointer{
			NodeId:   nodeId,
			Disk:     disk1.Id(),
			FileName: uuid.NewString(),
		},
		nodeId,
		putBlockId,
	)

	wr := <-resp
	if wr.Resp.Id != putBlockId {
		t.Error("Unknown put block id")
		return
	}

	putBlockId2 := model.PutBlockId(uuid.NewString())
	disk2.OutWrites <- model.NewWriteResultErr(
		"some error happened",
		nodeId,
		putBlockId2,
	)

	wr = <-resp
	if wr.Resp.Id != putBlockId2 {
		t.Error("Unknown put block id")
		return
	}

	putBlockId3 := model.PutBlockId(uuid.NewString())
	disk1.OutWrites <- model.NewWriteResultOk(
		model.DiskPointer{
			NodeId:   nodeId,
			Disk:     disk1.Id(),
			FileName: uuid.NewString(),
		},
		remoteNodeId,
		putBlockId3,
	)

	payload := <-sends
	if payload.Payload.Type() != model.SaveToDiskResp {
		t.Error("unknown send payload")
		return
	}
}
