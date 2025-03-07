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
	"io/fs"
	"path/filepath"
	"tealfs/pkg/disk"
	"tealfs/pkg/model"
	"testing"
)

func TestWriteData(t *testing.T) {
	f, path, nodeId, mgrDiskWrites, _, diskMgrWrites, _, _ := newDiskService()
	blockId := model.NewBlockId()
	data := []byte{0, 1, 2, 3, 4, 5}
	expectedPath := filepath.Join(path.String(), string(blockId))
	req := model.NewWriteRequest(
		nodeId,
		model.RawData{
			Ptr: model.DiskPointer{
				NodeId:   nodeId,
				FileName: string(blockId),
			},
			Data: data,
		},
	)
	mgrDiskWrites <- req
	result := <-diskMgrWrites
	if !result.Ok() {
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
	f, path, _, _, mgrDiskReads, _, diskMgrReads, _ := newDiskService()
	blockId := model.NewBlockId()
	caller := model.NewNodeId()
	data := []byte{0, 1, 2, 3, 4, 5}
	expectedPath := filepath.Join(path.String(), string(blockId))
	_ = f.WriteFile(expectedPath, data)
	req := model.NewReadRequest(
		caller,
		[]model.DiskPointer{{NodeId: "node1", FileName: string(blockId)}},
		blockId,
		model.GetBlockId("getBlockId1"),
	)
	mgrDiskReads <- req
	result := <-diskMgrReads
	if !result.Ok() {
		t.Error("Bad write result")
		return
	}
	if !bytes.Equal(result.Data().Data, data) {
		t.Error("Read data is wrong")
		return
	}
}

func TestReadNewFile(t *testing.T) {
	f, path, _, _, mgrDiskReads, _, diskMgrReads, _ := newDiskService()
	blockId := model.NewBlockId()
	caller := model.NewNodeId()
	data := []byte{0, 1, 2, 3, 4, 5}
	f.ReadError = fs.ErrNotExist
	expectedPath := filepath.Join(path.String(), string(blockId))
	_ = f.WriteFile(expectedPath, data)
	req := model.NewReadRequest(
		caller,
		[]model.DiskPointer{{NodeId: "node1", FileName: string(blockId)}},
		blockId,
		model.GetBlockId("getBlockId1"),
	)
	mgrDiskReads <- req
	result := <-diskMgrReads
	if !result.Ok() {
		t.Error("Bad write result")
		return
	}
	if !bytes.Equal(result.Data().Data, []byte{}) {
		t.Error("Written data is wrong")
		return
	}
}

func newDiskService() (*disk.MockFileOps, disk.Path, model.NodeId, chan model.WriteRequest, chan model.ReadRequest, chan model.WriteResult, chan model.ReadResult, disk.Disk) {
	f := disk.MockFileOps{}
	path := disk.NewPath("/some/fake/path", &f)
	id := model.NewNodeId()
	mgrDiskWrites := make(chan model.WriteRequest)
	mgrDiskReads := make(chan model.ReadRequest)
	diskMgrWrites := make(chan model.WriteResult)
	diskMgrReads := make(chan model.ReadResult)
	d := disk.New(path, id, mgrDiskWrites, mgrDiskReads, diskMgrWrites, diskMgrReads)
	return &f, path, id, mgrDiskWrites, mgrDiskReads, diskMgrWrites, diskMgrReads, d
}
