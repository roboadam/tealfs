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

package proto

import (
	"bytes"
	"tealfs/pkg/nodes"
	"tealfs/pkg/store"
)

type ReadRequest struct {
	Caller  nodes.Id
	BlockId store.Id
}

func (r *ReadRequest) ToBytes() []byte {
	callerId := StringToBytes(string(r.Caller))
	blockId := StringToBytes(string(r.BlockId))
	return AddType(ReadDataType, bytes.Join([][]byte{callerId, blockId}, []byte{}))
}

func ToReadRequest(data []byte) *ReadRequest {
	callerId, remainder := StringFromBytes(data)
	blockId, _ := StringFromBytes(remainder)
	rq := ReadRequest{
		Caller:  nodes.Id(callerId),
		BlockId: store.Id(blockId),
	}
	return &rq
}
