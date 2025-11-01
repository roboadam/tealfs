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

package disk_test

import (
	"bytes"
	"context"
	"io/fs"
	"path/filepath"
	"tealfs/pkg/disk"
	"tealfs/pkg/model"
	"testing"

	"github.com/google/uuid"
)

func TestWriteData(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	f, path, nodeId, mgrDiskWrites, _, diskMgrWrites, _, _ := newDiskService(ctx)
	blockId := model.NewBlockId()
	data := []byte{0, 1, 2, 3, 4, 5}
	expectedPath := filepath.Join(path.String(), string(blockId))
	req := model.WriteRequest{
		Caller: nodeId,
		Data: model.RawData{
			Ptr:  model.DiskPointer{NodeId: nodeId, Disk: "disk1", FileName: string(blockId)},
			Data: data,
		},
		ReqId: "putBlockId",
	}
	mgrDiskWrites <- req
	result := <-diskMgrWrites
	if !result.Ok {
		t.Error("Bad write result")
		return
	}
	if writtenData, err := f.ReadFile(expectedPath); err == nil {
		if !bytes.Equal(writtenData, data) {
			t.Error("Written data is wrong")
			return
		}
	} else {
		t.Error("Written path is wrong", err)
		return
	}
}

func TestReadData(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	f, path, _, _, mgrDiskReads, _, diskMgrReads, _ := newDiskService(ctx)
	blockId := model.NewBlockId()
	caller := model.NewNodeId()
	data := []byte{0, 1, 2, 3, 4, 5}
	expectedPath := filepath.Join(path.String(), string(blockId))
	_ = f.WriteFile(expectedPath, data)
	req := model.ReadRequest{
		Caller:  caller,
		Ptrs:    []model.DiskPointer{{NodeId: "node1", Disk: "disk1", FileName: string(blockId)}},
		BlockId: blockId,
		ReqId:   model.GetBlockId("getBlockId1"),
	}
	mgrDiskReads <- req
	result := <-diskMgrReads
	if !result.Ok {
		t.Error("Bad write result")
		return
	}
	if !bytes.Equal(result.Data.Data, data) {
		t.Error("Read data is wrong")
		return
	}
}

func TestReadNewFile(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	f, path, _, _, mgrDiskReads, _, diskMgrReads, _ := newDiskService(ctx)
	blockId := model.NewBlockId()
	caller := model.NewNodeId()
	data := []byte{0, 1, 2, 3, 4, 5}
	f.ReadError = fs.ErrNotExist
	expectedPath := filepath.Join(path.String(), string(blockId))
	_ = f.WriteFile(expectedPath, data)
	req := model.ReadRequest{
		Caller:  caller,
		Ptrs:    []model.DiskPointer{{NodeId: "node1", Disk: "disk1", FileName: string(blockId)}},
		BlockId: blockId,
		ReqId:   model.GetBlockId("getBlockId1"),
	}
	mgrDiskReads <- req
	result := <-diskMgrReads
	if !result.Ok {
		t.Error("Bad write result")
		return
	}
	if !bytes.Equal(result.Data.Data, []byte{}) {
		t.Error("Written data is wrong")
		return
	}
}

func TestGet(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	f := disk.MockFileOps{}
	path := disk.NewPath("/some/fake/path", &f)
	id := model.NewNodeId()
	diskId := model.DiskId(uuid.New().String())
	d := disk.New(path, id, diskId, ctx)

	_, ok := d.Get("blockId")
	if ok {
		t.Error("should be no block")
	}

	path.Save(model.RawData{
		Ptr:  model.DiskPointer{
			NodeId:   id,
			Disk:     diskId,
			FileName: "blockId",
		},
		Data: []byte{1,2,3},
	})

	data, ok := d.Get("blockId")
	if !ok {
		t.Error("should be a block")
	}

	if !bytes.Equal(data, []byte{1,2,3}) {
		t.Error("wrong data")
	}
}

func newDiskService(ctx context.Context) (*disk.MockFileOps, disk.Path, model.NodeId, chan model.WriteRequest, chan model.ReadRequest, chan model.WriteResult, chan model.ReadResult, disk.Disk) {
	f := disk.MockFileOps{}
	path := disk.NewPath("/some/fake/path", &f)
	id := model.NewNodeId()
	diskId := model.DiskId(uuid.New().String())
	d := disk.New(path, id, diskId, ctx)
	return &f, path, id, d.InWrites, d.InReads, d.OutWrites, d.OutReads, d
}
