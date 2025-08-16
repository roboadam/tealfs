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
	"tealfs/pkg/disk"
	"tealfs/pkg/model"
	"tealfs/pkg/set"
	"testing"

	"github.com/google/uuid"
)

func TestLocalBlockReader(t *testing.T) {
	nodeId := model.NewNodeId()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	req := make(chan GetFromDiskReq)

	disk1, blockId1 := mockDisk(nodeId, ctx)
	disk2, _ := mockDisk(nodeId, ctx)
	disks := set.NewSet[disk.Disk]()
	disks.Add(*disk1)
	disks.Add(*disk2)

	lbs := LocalBlockReader{
		Req:   req,
		Disks: &disks,
	}

	go lbs.Start(ctx)

	blockReq := model.NewGetBlockReq(blockId1)
	req <- GetFromDiskReq{
		Caller: model.NewNodeId(),
		Dest: Dest{
			NodeId: nodeId,
			DiskId: disk1.Id(),
		},
		Req: blockReq,
	}

	result := <-disk1.OutReads
	if !result.Ok {
		t.Errorf("Expected read to succeed, got error: %s", result.Message)
		return
	}

	if result.ReqId != blockReq.Id {
		t.Errorf("Expected different block id")
	}
}

func mockDisk(nodeId model.NodeId, ctx context.Context) (*disk.Disk, model.BlockId) {
	p := disk.NewPath("/test", &disk.MockFileOps{})
	d := disk.New(
		p,
		nodeId,
		model.DiskId(uuid.NewString()),
		ctx,
	)
	blockId := model.NewBlockId()
	d.InWrites <- model.WriteRequest{
		Caller: "",
		Data: model.RawData{
			Ptr: model.DiskPointer{
				NodeId:   nodeId,
				Disk:     d.Id(),
				FileName: string(blockId),
			},
			Data: []byte{1, 2, 3, 4, 5},
		},
		ReqId: "",
	}
	<-d.OutWrites
	return &d, blockId
}
