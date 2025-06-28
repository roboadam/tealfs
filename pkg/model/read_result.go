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
	blockId BlockId
	reqId   GetBlockId
}

func NewReadResultOk(
	caller NodeId,
	ptrs []DiskPointer,
	data RawData,
	reqId GetBlockId,
	blockId BlockId,
) ReadResult {
	return ReadResult{
		ok:      true,
		message: "",
		caller:  caller,
		ptrs:    ptrs,
		data:    data,
		reqId:   reqId,
		blockId: blockId,
	}
}

func NewReadResultErr(
	message string,
	caller NodeId,
	reqId GetBlockId,
	blockId BlockId,
) ReadResult {
	return ReadResult{
		ok:      false,
		message: message,
		caller:  caller,
		reqId:   reqId,
		blockId: blockId,
	}
}

func (r *ReadResult) Ok() bool            { return r.ok }
func (r *ReadResult) Message() string     { return r.message }
func (r *ReadResult) Caller() NodeId      { return r.caller }
func (r *ReadResult) Ptrs() []DiskPointer { return r.ptrs }
func (r *ReadResult) Data() RawData       { return r.data }
func (r *ReadResult) ReqId() GetBlockId   { return r.reqId }
func (r *ReadResult) BlockId() BlockId    { return r.blockId }

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
	blockId := StringToBytes(string(r.blockId))
	payload := bytes.Join([][]byte{ok, message, caller, numPtrs, ptrs, raw, reqId, blockId}, []byte{})
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
	reqId, remainder := StringFromBytes(remainder)
	blockId, _ := StringFromBytes(remainder)
	return &ReadResult{
		ok:      ok,
		message: message,
		caller:  NodeId(caller),
		ptrs:    ptrs,
		data:    *raw,
		reqId:   GetBlockId(reqId),
		blockId: BlockId(blockId),
	}
}
