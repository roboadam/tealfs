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

package datalayer_test

import (
	"context"
	"fmt"
	"tealfs/pkg/datalayer"
	"tealfs/pkg/disk"
	"tealfs/pkg/model"
	"tealfs/pkg/set"
	"testing"
)

func TestDeleteRequestHandlerDeletesBlock(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	dataRemoved := make(chan struct{}, 1)
	fileOps := disk.MockFileOps{DataRemoved: dataRemoved}
	disks := set.NewSet[disk.Disk]()
	nodeId := model.NodeId("nodeId")
	diskId := model.DiskId("disk1")
	path := "somePath"
	blockId := model.BlockId("blockId")
	fileData := []byte{1, 2, 3, 4, 5}

	fileOps.WriteFile(fmt.Sprint(path, "/", blockId), fileData)

	d := disk.New(disk.NewPath(path, &fileOps), nodeId, diskId, ctx)
	disks.Add(d)

	drh := datalayer.DeleteRequestHandler{
		Disks:        &disks,
		NodeId:       nodeId,
		StateHandler: nil,
	}

	req := datalayer.DeleteRequest{
		Dest:    model.NodeDisk{DiskId: diskId, NodeId: nodeId},
		BlockId: blockId,
	}

	drh.HandleDeleteRequest(&req)

	<-dataRemoved

	if fileOps.Exists(fmt.Sprint(path, "/", blockId)) {
		t.Fatal("file should have been deleted")
	}
}

func TestDeleteRequestHandlerIgnoresUnknownDisk(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	fileOps := disk.MockFileOps{}
	disks := set.NewSet[disk.Disk]()
	nodeId := model.NodeId("nodeId")
	diskId := model.DiskId("disk1")
	unknownDiskId := model.DiskId("unknownDisk")
	path := "somePath"
	blockId := model.BlockId("blockId")
	fileData := []byte{1, 2, 3, 4, 5}
	filePath := fmt.Sprint(path, "/", blockId)

	fileOps.WriteFile(filePath, fileData)

	d := disk.New(disk.NewPath(path, &fileOps), nodeId, diskId, ctx)
	disks.Add(d)

	drh := datalayer.DeleteRequestHandler{
		Disks:        &disks,
		NodeId:       nodeId,
		StateHandler: nil,
	}

	dataRemoved := make(chan struct{}, 1)
	fileOps2 := disk.MockFileOps{DataRemoved: dataRemoved}
	diskId2 := model.DiskId("disk2")
	path2 := "otherPath"
	blockId2 := model.BlockId("blockId2")
	fileOps2.WriteFile(fmt.Sprint(path2, "/", blockId2), fileData)
	d2 := disk.New(disk.NewPath(path2, &fileOps2), nodeId, diskId2, ctx)
	disks.Add(d2)

	req1 := datalayer.DeleteRequest{
		Dest:    model.NodeDisk{DiskId: unknownDiskId, NodeId: nodeId},
		BlockId: blockId,
	}

	req2 := datalayer.DeleteRequest{
		Dest:    model.NodeDisk{DiskId: diskId2, NodeId: nodeId},
		BlockId: blockId2,
	}

	drh.HandleDeleteRequest(&req1)
	drh.HandleDeleteRequest(&req2)

	<-dataRemoved

	if !fileOps.Exists(filePath) {
		t.Fatal("file on unknown disk should not have been deleted")
	}
}
