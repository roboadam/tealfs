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
	wr := model.WriteRequest{
		Caller: "caller1",
		Block: model.Block{
			Key: model.BlockKey{
				Id:   "block1",
				Type: model.XORed,
				Data: []model.DiskPointer{
					{
						NodeId:   "node1",
						FileName: "file1",
					},
				},
				Parity: model.DiskPointer{
					NodeId:   "node2",
					FileName: "file2",
				},
			},
			Data: []byte{0x01, 0x02, 0x03},
		},
	}
	raw := wr.ToBytes()
	newWr := model.ToWriteRequest(raw[1:])
	if !wr.Equal(newWr) {
		t.Errorf("Expected %v, got %v", wr, newWr)
	}
}
