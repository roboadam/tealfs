// Copyright (C) 2024 Adam Hess
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
	"encoding/hex"
	"path/filepath"
	"tealfs/pkg/disk"
	"tealfs/pkg/hash"
	"tealfs/pkg/model"
	"testing"
)

func TestWriteData(t *testing.T) {
	f := disk.MockFileOps{}
	path := disk.NewPath("/some/fake/path", &f)
	id := model.NewNodeId()
	mgrDiskWrites := make(chan model.Block)
	mgrDiskReads := make(chan model.ReadRequest)
	diskMgrWrites := make(chan model.WriteResult)
	diskMgrReads := make(chan model.ReadResult)
	blockId := model.NewBlockId()
	data := []byte{0, 1, 2, 3, 4, 5}
	h := hash.ForData(data)
	expectedPath := filepath.Join(path.String(), hex.EncodeToString(h.Value))
	_ = disk.New(path, id, mgrDiskWrites, mgrDiskReads, diskMgrWrites, diskMgrReads)
	mgrDiskWrites <- model.Block{
		Id:   blockId,
		Data: data,
		Hash: h,
	}
	result := <-diskMgrWrites
	if !result.Ok {
		t.Error("Bad write result")
	}
	if !bytes.Equal(f.WrittenData, data) {
		t.Error("Written data is wrong")
	}
	if f.WritePath != expectedPath {
		t.Error("Written path is wrong")
	}
}

func TestReadData(t *testing.T) {
	f := disk.MockFileOps{}
	path := disk.NewPath("/some/fake/path", &f)
	id := model.NewNodeId()
	mgrDiskWrites := make(chan model.Block)
	mgrDiskReads := make(chan model.ReadRequest)
	diskMgrWrites := make(chan model.WriteResult)
	diskMgrReads := make(chan model.ReadResult)
	blockId := model.NewBlockId()
	data := []byte{0, 1, 2, 3, 4, 5}
	h := hash.ForData(data)
	expectedPath := filepath.Join(path.String(), hex.EncodeToString(h.Value))
	_ = disk.New(path, id, mgrDiskWrites, mgrDiskReads, diskMgrWrites, diskMgrReads)
	mgrDiskWrites <- model.Block{
		Id:   blockId,
		Data: data,
		Hash: h,
	}
	result := <-diskMgrWrites
	if !result.Ok {
		t.Error("Bad write result")
	}
	if !bytes.Equal(f.WrittenData, data) {
		t.Error("Written data is wrong")
	}
	if f.WritePath != expectedPath {
		t.Error("Written path is wrong")
	}
}

func NewDiskService() (disk.MockFileOps, disk.Path, model.NodeId, chan model.Block, chan model.ReadRequest, chan model.WriteResult, chan model.ReadResult, disk.Disk ) {
	f := disk.MockFileOps{}
	path := disk.NewPath("/some/fake/path", &f)
	id := model.NewNodeId()
	mgrDiskWrites := make(chan model.Block)
	mgrDiskReads := make(chan model.ReadRequest)
	diskMgrWrites := make(chan model.WriteResult)
	diskMgrReads := make(chan model.ReadResult)
	d := disk.New(path, id, mgrDiskWrites, mgrDiskReads, diskMgrWrites, diskMgrReads)
	return f, path, id, mgrDiskWrites, mgrDiskReads, diskMgrWrites, diskMgrReads, d
}