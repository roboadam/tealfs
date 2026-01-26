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

package disk

import (
	"context"
	"tealfs/pkg/model"
	"tealfs/pkg/set"
	"testing"
	"time"
)

func TestDisks(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	inAddDiskMsg := make(chan model.AddDiskMsg)
	inDiskAddedMsg := make(chan model.DiskAddedMsg)
	outDiskAddedMsg := make(chan model.DiskAddedMsg)
	outReadResult := make(chan (<-chan model.ReadResult))
	outWriteResult := make(chan (<-chan model.WriteResult))
	fileOps := MockFileOps{}

	diskMgrSvc := NewDisks("localNodeId", "", &fileOps)
	diskMgrSvc.InAddDiskMsg = inAddDiskMsg
	diskMgrSvc.InDiskAddedMsg = inDiskAddedMsg
	diskMgrSvc.OutDiskAddedMsg = outDiskAddedMsg
	diskMgrSvc.OutAddedReadResults = outReadResult
	diskMgrSvc.OutAddedWriteResults = outWriteResult
	go diskMgrSvc.Start(ctx)

	localDisk := model.AddDiskMsg{
		DiskId: "localDisk1",
		Path:   "localPath",
		NodeId: "localNodeId",
	}

	inAddDiskMsg <- localDisk
	<-outDiskAddedMsg
	<-outReadResult
	<-outWriteResult

	for diskMgrSvc.DiskInfoList.Len() != 1 {
		time.Sleep(time.Millisecond)
	}

	for diskMgrSvc.LocalDiskSvcList.Len() != 1 {
		time.Sleep(time.Millisecond)
	}

	inDiskAddedMsg <- model.DiskAddedMsg{
		DiskId: "disk1",
		Path:   "path1",
		NodeId: "remoteNodeId",
	}

	for diskMgrSvc.DiskInfoList.Len() != 2 {
		time.Sleep(time.Millisecond)
	}

	for diskMgrSvc.LocalDiskSvcList.Len() != 1 {
		time.Sleep(time.Millisecond)
	}

	cancel()

	diskMgrSvc = NewDisks("localNodeId", "", &fileOps)
	diskMgrSvc.InAddDiskMsg = inAddDiskMsg
	diskMgrSvc.InDiskAddedMsg = inDiskAddedMsg
	diskMgrSvc.OutDiskAddedMsg = outDiskAddedMsg
	go diskMgrSvc.Start(ctx)

	for diskMgrSvc.DiskInfoList.Len() != 2 {
		time.Sleep(time.Millisecond)
	}

	for diskMgrSvc.LocalDiskSvcList.Len() != 1 {
		time.Sleep(time.Millisecond)
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

func TestListIds(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	fileOps := MockFileOps{}
	path := NewPath("", &fileOps)
	disk := New(path, "nodeId", "diskId", ctx)

	req := ListIds{
		Resp: make(chan set.Set[model.BlockId], 1),
	}

	disk.InListIds <- req
	list := <-req.Resp

	if list.Len() != 0 {
		t.Error("invalid response")
	}

	fileOps.WriteFile("blockId", []byte{1, 2, 3, 4, 5})

	disk.InListIds <- req
	list = <-req.Resp

	if list.Len() != 1 {
		t.Error("invalid response")
	}
}
