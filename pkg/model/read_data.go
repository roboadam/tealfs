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

type ReadRequest struct {
	Caller   NodeId
	BlockKey BlockKey
}

func (r *ReadRequest) ToBytes() []byte {
	callerId := StringToBytes(string(r.Caller))
	blockKey := r.BlockKey.ToBytes()
	return AddType(ReadRequestType, bytes.Join([][]byte{callerId, blockKey}, []byte{}))
}

func (r *ReadRequest) Equal(p Payload) bool {
	if o, ok := p.(*ReadRequest); ok {
		return r.Caller == o.Caller && r.BlockKey.Equals(&o.BlockKey)
	}
	return false
}

func ToReadRequest(data []byte) *ReadRequest {
	callerId, remainder := StringFromBytes(data)
	blockKey, _ := ToBlockKey(remainder)
	rq := ReadRequest{
		Caller:   NodeId(callerId),
		BlockKey: *blockKey,
	}
	return &rq
}
