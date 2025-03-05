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
	ok      bool
	message string
	caller  NodeId
	ptrs    []DiskPointer
	data    RawData
	reqId   GetBlockId
}

func NewReadResultOk(
	caller NodeId,
	ptrs []DiskPointer,
	data RawData,
	reqId GetBlockId,
) ReadResult {
	return ReadResult{
		ok:      true,
		message: "",
		caller:  caller,
		ptrs:    ptrs,
		data:    data,
		reqId:   reqId,
	}
}

func NewReadResultErr(
	message string,
	caller NodeId,
	reqId GetBlockId,
) ReadResult {
	return ReadResult{
		ok:      false,
		message: message,
		caller:  caller,
		reqId:   reqId,
	}
}

func (r *ReadResult) Equal(p Payload) bool {

	if o, ok := p.(*ReadResult); ok {
		if r.ok != o.ok {
			return false
		}
		if r.message != o.message {
			return false
		}
		if r.caller != o.caller {
			return false
		}
		if len(r.ptrs) != len(o.ptrs) {
			return false
		}
		for i, ptr := range r.ptrs {
			if !ptr.Equals(&o.ptrs[i]) {
				return false
			}
		}
		if !r.data.Equals(&o.data) {
			return false
		}
		if r.reqId != o.reqId {
			return false
		}

		return true
	}
	return false
}

func (r *ReadResult) ToBytes() []byte {
	ok := BoolToBytes(r.ok)
	message := StringToBytes(r.message)
	caller := StringToBytes(string(r.caller))
	numPtrs := IntToBytes(uint32(len(r.ptrs)))
	ptrs := make([]byte, 0)
	for _, ptr := range r.ptrs {
		ptrs = append(ptrs, ptr.ToBytes()...)
	}
	raw := r.data.ToBytes()
	reqId := StringToBytes(string(r.reqId))
	payload := bytes.Join([][]byte{ok, message, caller, numPtrs, ptrs, raw, reqId}, []byte{})
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
	reqId, _ := StringFromBytes(remainder)
	return &ReadResult{
		ok:      ok,
		message: message,
		caller:  NodeId(caller),
		ptrs:    ptrs,
		data:    *raw,
		reqId:   GetBlockId(reqId),
	}
}
