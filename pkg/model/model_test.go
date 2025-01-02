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
	rr1 := model.ReadResult{
		Ok:      true,
		Message: "some message",
		Caller:  "some caller",
		Block: model.Block{
			Key: model.BlockKey{
				Id:   "blockKeyId",
				Type: model.Mirrored,
				Data: []model.DiskPointer{{
					NodeId:   "nodeId",
					FileName: "fileName",
				}},
			},
			Data: []byte{1, 2, 3},
		},
	}
	rr2 := model.ReadResult{
		Ok:      false,
		Message: "some message",
		Caller:  "some caller",
		Block: model.Block{
			Key: model.BlockKey{
				Id:   "blockKeyId",
				Type: model.Mirrored,
				Data: []model.DiskPointer{{
					NodeId:   "nodeId",
					FileName: "fileName",
				}},
			},
			Data: []byte{1, 2, 3},
		},
	}
	if rr1.Equal(&rr2) {
		t.Error("should not be equal")
	}

	rr2.Ok = true

	if !rr1.Equal(&rr2) {
		t.Error("should be equal")
	}

	bytes1 := rr1.ToBytes()
	rr3 := model.ToReadResult(bytes1[1:])

	if !rr1.Equal(rr3) {
		t.Error("should be equal")
	}
}

func TestWriteResult(t *testing.T) {
	wr1 := model.WriteResult{
		Ok:      true,
		Message: "some message",
		Caller:  "some caller",
		BlockKey: model.BlockKey{
			Id:   "blockKeyId",
			Type: model.Mirrored,
			Data: []model.DiskPointer{{
				NodeId:   "nodeId",
				FileName: "fileName",
			}},
		},
	}
	wr2 := model.WriteResult{
		Ok:      true,
		Message: "some message",
		Caller:  "some caller 2",
		BlockKey: model.BlockKey{
			Id:   "blockKeyId",
			Type: model.Mirrored,
			Data: []model.DiskPointer{{
				NodeId:   "nodeId",
				FileName: "fileName",
			}},
		},
	}

	if wr1.Equal(&wr2) {
		t.Error("should not be equal")
	}

	wr2.Caller = "some caller"

	if !wr1.Equal(&wr2) {
		t.Error("should be equal")
	}

	bytes1 := wr1.ToBytes()
	rr3 := model.ToWriteResult(bytes1[1:])

	if !wr1.Equal(rr3) {
		t.Error("should be equal")
	}
}
