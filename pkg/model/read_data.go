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

package model

import (
	"bytes"
)

type ReadRequest struct {
	Caller  NodeId
	Ptrs    []DiskPointer
	BlockId BlockId
}

func (r *ReadRequest) ToBytes() []byte {
	callerId := StringToBytes(string(r.Caller))
	ptrLen := IntToBytes(uint32(len(r.Ptrs)))
	ptrs := make([]byte, 0)
	for _, ptr := range r.Ptrs {
		ptrs = append(ptrs, ptr.ToBytes()...)
	}
	blockId := StringToBytes(string(r.BlockId))
	return AddType(ReadRequestType, bytes.Join([][]byte{callerId, ptrLen, ptrs, blockId}, []byte{}))
}

func (r *ReadRequest) Equal(p Payload) bool {
	if o, ok := p.(*ReadRequest); ok {
		if r.Caller != o.Caller {
			return false
		}
		if len(r.Ptrs) != len(o.Ptrs) {
			return false
		}
		for i, ptr := range r.Ptrs {
			if !ptr.Equals(&o.Ptrs[i]) {
				return false
			}
		}
		return r.BlockId != o.BlockId
	}
	return false
}

func ToReadRequest(data []byte) *ReadRequest {
	callerId, remainder := StringFromBytes(data)
	numPtrs, remainder := IntFromBytes(remainder)
	ptrs := make([]DiskPointer, numPtrs)
	for range numPtrs {
		var ptr *DiskPointer
		ptr, remainder = ToDiskPointer(remainder)
		ptrs = append(ptrs, *ptr)
	}
	blockId, _ := StringFromBytes(remainder)
	rq := ReadRequest{
		Caller:  NodeId(callerId),
		Ptrs:    ptrs,
		BlockId: BlockId(blockId),
	}
	return &rq
}
