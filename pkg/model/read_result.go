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

type ReadResult struct {
	Ok      bool
	Message string
	Caller  NodeId
	Ptrs    []DiskPointer
	Data    RawData
	BlockId BlockId
}

func (r *ReadResult) Equal(p Payload) bool {

	if o, ok := p.(*ReadResult); ok {
		if r.Ok != o.Ok {
			return false
		}
		if r.Message != o.Message {
			return false
		}
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
		if !r.Data.Equals(&o.Data) {
			return false
		}
		if r.BlockId != o.BlockId {
			return false
		}

		return true
	}
	return false
}

func (r *ReadResult) ToBytes() []byte {
	ok := BoolToBytes(r.Ok)
	message := StringToBytes(r.Message)
	caller := StringToBytes(string(r.Caller))
	numPtrs := IntToBytes(uint32(len(r.Ptrs)))
	ptrs := make([]byte, 0)
	for _, ptr := range r.Ptrs {
		ptrs = append(ptrs, ptr.ToBytes()...)
	}
	raw := r.Data.ToBytes()
	blockId := StringToBytes(string(r.BlockId))
	payload := bytes.Join([][]byte{ok, message, caller, numPtrs, ptrs, raw, blockId}, []byte{})
	return AddType(ReadResultType, payload)
}

func ToReadResult(data []byte) *ReadResult {
	ok, remainder := BoolFromBytes(data)
	message, remainder := StringFromBytes(remainder)
	caller, remainder := StringFromBytes(remainder)
	numPtrs, remainder := IntFromBytes(remainder)
	ptrs := make([]DiskPointer, 0, numPtrs)
	for range numPtrs {
		var ptr *DiskPointer
		ptr, remainder = ToDiskPointer(remainder)
		ptrs = append(ptrs, *ptr)
	}
	raw, remainder := ToRawData(remainder)
	blockId, _ := StringFromBytes(remainder)
	return &ReadResult{
		Ok:      ok,
		Message: message,
		Caller:  NodeId(caller),
		Ptrs:    ptrs,
		Data:    *raw,
		BlockId: BlockId(blockId),
	}
}
