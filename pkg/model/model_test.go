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

package model_test

import (
	"tealfs/pkg/model"
	"testing"
)

func TestSyncNodes(t *testing.T) {
	n1 := struct {
		Node    model.NodeId
		Address string
	}{
		Node:    model.NewNodeId(),
		Address: "node:1",
	}
	n2 := struct {
		Node    model.NodeId
		Address string
	}{
		Node:    model.NewNodeId(),
		Address: "node:2",
	}
	sn1 := model.NewSyncNodes()
	sn1.Nodes.Add(n1)
	sn1.Nodes.Add(n2)
	sn2 := model.NewSyncNodes()
	sn2.Nodes.Add(n2)
	sn2.Nodes.Add(n1)

	if !sn1.Equal(&sn2) {
		t.Error("should be equal")
	}

	bytes1 := sn1.ToBytes()
	sn3 := model.ToSyncNodes(bytes1[1:])

	if !sn1.Equal(sn3) {
		t.Error("should be equal")
	}
}

func TestReadResult(t *testing.T) {
	caller := model.NodeId("some caller")
	ptrs := []model.DiskPointer{
		{NodeId: "node1", FileName: "fileName1"},
		{NodeId: "node2", FileName: "fileName2"},
	}
	data1 := model.RawData{
		Ptr:  model.DiskPointer{NodeId: "node1", FileName: "fileName1"},
		Data: []byte{1, 2, 3},
	}
	data2 := model.RawData{
		Ptr:  model.DiskPointer{NodeId: "node1", FileName: "fileName1"},
		Data: []byte{1, 2, 4},
	}
	reqId := model.GetBlockId("getBlockId")
	rr1 := model.NewReadResultOk(caller, ptrs, data1, reqId)
	rr2 := model.NewReadResultOk(caller, ptrs, data2, reqId)
	if rr1.Equal(&rr2) {
		t.Error("should not be equal")
	}

	bytes1 := rr1.ToBytes()
	rr3 := model.ToReadResult(bytes1[1:])

	if !rr1.Equal(rr3) {
		t.Error("should be equal")
	}
}

func TestWriteResult(t *testing.T) {
	caller := model.NodeId("some caller")
	ptr := model.DiskPointer{NodeId: "nodeId", FileName: "fileName"}
	reqId1 := model.PutBlockId("reqId1")
	reqId2 := model.PutBlockId("reqId2")
	wr1 := model.NewWriteResultOk(ptr, caller, reqId1)
	wr2 := model.NewWriteResultOk(ptr, caller, reqId2)
	if wr1.Equal(&wr2) {
		t.Error("should not be equal")
		return
	}

	bytes1 := wr1.ToBytes()
	rr3 := model.ToWriteResult(bytes1[1:])

	if !wr1.Equal(rr3) {
		t.Error("should be equal")
		return
	}
}

func TestReadRequest(t *testing.T) {
	caller := model.NodeId("caller1")
	ptrs := []model.DiskPointer{{NodeId: "nodeId1", FileName: "filename1"}}
	blockId := model.BlockId("blockId1")
	rr1 := model.NewReadRequest(caller, ptrs, blockId)
	rr2 := model.NewReadRequest(caller, ptrs, blockId)

	if rr1.Equal(&rr2) {
		t.Error("should not be equal because of the internal request id")
		return
	}

	bytes1 := rr1.ToBytes()
	rr3 := model.ToReadRequest(bytes1[1:])

	if !rr1.Equal(rr3) {
		t.Error("should be equal")
		return
	}
}
