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

package model

import (
	"bytes"
)

type WriteRequest struct {
	Caller Id
	Block  Block
}

func (r *WriteRequest) Equal(p Payload) bool {
	if o, ok := p.(*WriteRequest); ok {
		if r.Caller != o.Caller {
			return false
		}
		if r.Block.Equal(&o.Block) {
			return false
		}
		return true
	}
	return false
}

func (r *WriteRequest) ToBytes() []byte {
	caller := StringToBytes(string(r.Caller))
	block := BlockToBytes(r.Block)
	return bytes.Join([][]byte{caller, block}, []byte{})
}

func ToWriteRequest(data []byte) *WriteRequest {
	caller, remainder := StringFromBytes(data)
	block, _ := BlockFromBytes(remainder)
	return &WriteRequest{
		Caller: Id(caller),
		Block:  block,
	}
}
