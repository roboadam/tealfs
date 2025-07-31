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

package disk

import (
	"context"
	"tealfs/pkg/disk/dist"
	"tealfs/pkg/model"
	"tealfs/pkg/set"
	"testing"
)

func TestLocalDiskAdder(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	inAddDiskReq := make(chan model.AddDiskReq)
	diskChan1 := make(chan *Disk, 1)
	diskChan2 := make(chan *Disk, 1)
	outAddLocalDisk := []chan<- *Disk{diskChan1, diskChan2}
	outIamDiskUpdate := make(chan []model.AddDiskReq, 1)

	nodeId := model.NewNodeId()
	fileOps := MockFileOps{}
	disks := set.NewSet[Disk]()
	distributer := dist.NewMirrorDistributer(nodeId)
	allDiskIds := set.NewSet[model.AddDiskReq]()

	adder := LocalDiskAdder{
		InAddDiskReq:     inAddDiskReq,
		OutAddLocalDisk:  outAddLocalDisk,
		OutIamDiskUpdate: outIamDiskUpdate,

		FileOps:     &fileOps,
		Disks:       &disks,
		Distributer: &distributer,
		AllDiskIds:  &allDiskIds,
	}
	go adder.Start(ctx)

	inAddDiskReq <- model.AddDiskReq{
		DiskId: "diskId1",
		Path:   "path1",
		NodeId: nodeId,
	}

	disk1 := <-diskChan1
	disk2 := <-diskChan2
	iamUpdate := <- outIamDiskUpdate

	if disk1.diskId != "diskId1" || disk2.diskId != "diskId1" {
		t.Error("invalid disk")
	}
	if len(iamUpdate) != 1 {
		t.Error("invalid number of disks")
	}

	inAddDiskReq <- model.AddDiskReq{
		DiskId: "diskId2",
		Path:   "path2",
		NodeId: nodeId,
	}

	iamUpdate = <- outIamDiskUpdate

	if len(iamUpdate) != 2 {
		t.Error("invalid number of disks")
	}
}
