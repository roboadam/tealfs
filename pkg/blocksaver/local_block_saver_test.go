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
	"tealfs/pkg/set"
	"testing"

	"github.com/google/uuid"
)

func TestLocalBlockSaver(t *testing.T) {
	nodeId := model.NewNodeId()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	req := make(chan SaveToDiskReq)

	disk1 := mockDisk(nodeId, ctx)
	disk2 := mockDisk(nodeId, ctx)
	disks := set.NewSet[disk.Disk]()
	disks.Add(*disk1)
	disks.Add(*disk2)

	lbs := LocalBlockSaver{
		Req:   req,
		Disks: &disks,
	}

	go lbs.Start(ctx)

	req <- SaveToDiskReq{
		Caller: model.NewNodeId(),
		Dest: Dest{
			NodeId: nodeId,
			DiskId: disk1.Id(),
		},
		Req: model.NewPutBlockReq(model.Block{
			Id:   model.NewBlockId(),
			Data: []byte{1, 2, 3, 4, 5, 6, 7, 8, 9, 10},
		}),
	}

	result := <-disk1.OutWrites
	if !result.Ok {
		t.Errorf("Expected write to succeed, got error: %s", result.Message)
	}

	if result.Ptr.NodeId != nodeId || result.Ptr.Disk != disk1.Id() {
		t.Errorf("Expected DiskPointer to match, got %v", result.Ptr)
	}

	req <- SaveToDiskReq{
		Caller: model.NewNodeId(),
		Dest: Dest{
			NodeId: nodeId,
			DiskId: disk2.Id(),
		},
		Req: model.NewPutBlockReq(model.Block{
			Id:   model.NewBlockId(),
			Data: []byte{10, 9, 8, 7, 6, 5, 4, 3, 2, 1},
		}),
	}
	result = <-disk2.OutWrites
	if !result.Ok {
		t.Errorf("Expected write to succeed, got error: %s", result.Message)
	}
	if result.Ptr.NodeId != nodeId || result.Ptr.Disk != disk2.Id() {
		t.Errorf("Expected DiskPointer to match, got %v", result.Ptr)
	}
}

func mockDisk(nodeId model.NodeId, ctx context.Context) *disk.Disk {
	p := disk.NewPath("/test", &disk.MockFileOps{})
	d := disk.New(
		p,
		nodeId,
		model.DiskId(uuid.NewString()),
		make(chan model.ReadRequest),
		make(chan model.ReadResult),
		ctx,
	)
	return &d
}
