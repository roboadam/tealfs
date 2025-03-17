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

func TestWriteRequest(t *testing.T) {
	caller := model.NodeId("caller1")
	data := model.RawData{
		Ptr:  model.NewDiskPointer("node1", "disk1", "fileName1"),
		Data: []byte{0x01, 0x02, 0x03},
	}
	wr := model.NewWriteRequest(caller, data, "putBlockId")

	raw := wr.ToBytes()
	newWr := model.ToWriteRequest(raw[1:])
	if !wr.Equal(newWr) {
		t.Errorf("Expected %v, got %v", wr, newWr)
	}
}
