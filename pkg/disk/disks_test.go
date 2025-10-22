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
	"reflect"
	"tealfs/pkg/model"
	"testing"

	"github.com/google/uuid"
)

func TestDisks(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	inAddDiskReq := make(chan model.AddDiskReq)
	outRemoteAddDiskReq := make(chan model.AddDiskReq)
	outLocalAddDiskReq := make(chan model.AddDiskReq)
	localNodeId := model.NewNodeId()
	remoteNodeId := model.NewNodeId()

	disks := NewDisks(localNodeId)
	disks.InAddDiskReq = inAddDiskReq
	disks.OutRemoteAddDiskReq = outRemoteAddDiskReq
	disks.OutLocalAddDiskReq = outLocalAddDiskReq
	go disks.Start(ctx)

	localDisk := model.AddDiskReq{
		DiskId: model.DiskId(uuid.NewString()),
		Path:   "localPath",
		NodeId: localNodeId,
	}
	remoteDisk := model.AddDiskReq{
		DiskId: model.DiskId(uuid.NewString()),
		Path:   "remotePath",
		NodeId: remoteNodeId,
	}

	inAddDiskReq <- localDisk
	localResp := <-outLocalAddDiskReq
	disks.AllDiskIds.Add(localResp)
	inAddDiskReq <- remoteDisk
	remoteResp := <-outRemoteAddDiskReq
	disks.AllDiskIds.Add(remoteResp)

	inAddDiskReq <- localDisk
	inAddDiskReq <- remoteDisk

	select {
	case <-outLocalAddDiskReq:
		t.Error("we already got this message")
		return
	case <-outRemoteAddDiskReq:
		t.Error("we already got this message")
		return
	default:
	}

	if !reflect.DeepEqual(localDisk, localResp) {
		t.Error("local not equal")
	}

	if !reflect.DeepEqual(remoteDisk, remoteResp) {
		t.Error("remote not equal")
	}
}

func TestExists(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	fileOps := MockFileOps{}
	path := NewPath("", &fileOps)
	disk := New(path, "nodeId", "diskId", ctx)
	resp := make(chan bool, 1)

	req := ExistsReq{
		BlockId: "blockId",
		Resp:    resp,
	}

	disk.InExists <- req
	exists := <-resp

	if exists {
		t.Error("invalid response")
	}

	fileOps.WriteFile("blockId", []byte{1, 2, 3, 4, 5})

	disk.InExists <- req
	exists = <-resp

	if !exists {
		t.Error("invalid response")
	}
}
