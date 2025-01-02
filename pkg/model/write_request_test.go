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

func TestWriteRequest(t *testing.T) {
	key := model.BlockKey{
		Id:   model.BlockKeyId("key"), fixme
		Type: model.Mirrored,
		Data: []model.DiskPointer{
			{NodeId: model.NodeId("nodeId1"), FileName: "fileName1"},
		},
		Parity: model.DiskPointer{NodeId: model.NodeId("nodeId2"), FileName: "fileName2"},
	}
	raw := key.ToBytes()
	newKey, remainder := model.ToBlockKey(raw)
	if !key.Equals(newKey) {
		t.Errorf("Expected %v, got %v", key, newKey)
	}
	if len(remainder) != 0 {
		t.Errorf("Expected no remainder, got %v", remainder)
	}
}
