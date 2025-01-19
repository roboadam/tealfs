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

func TestDiskPtr(t *testing.T) {
	ptr := model.DiskPointer{
		NodeId:   model.NodeId("nodeId"),
		FileName: "fileName",
	}
	raw := ptr.ToBytes()
	newPtr, remainder := model.ToDiskPointer(raw)
	if !ptr.Equals(newPtr) {
		t.Errorf("Expected %v, got %v", ptr, newPtr)
	}
	if len(remainder) != 0 {
		t.Errorf("Expected no remainder, got %v", remainder)
	}
}
