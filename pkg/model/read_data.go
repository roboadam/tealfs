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
	caller  NodeId
	ptrs    []DiskPointer
	blockId BlockId
	reqId   GetBlockId
}

func NewReadRequest(
	caller NodeId,
	ptrs []DiskPointer,
	blockId BlockId,
	reqId GetBlockId,
) ReadRequest {
	return ReadRequest{
		caller:  caller,
		ptrs:    ptrs,
		blockId: blockId,
		reqId:   reqId,
	}
}

func (r *ReadRequest) Caller() NodeId {
	return r.caller
}
func (r *ReadRequest) Ptrs() []DiskPointer {
	return r.ptrs
}
func (r *ReadRequest) BlockId() BlockId {
	return r.blockId
}
func (r *ReadRequest) GetBlockId() GetBlockId {
	return r.reqId
}

func (r *ReadRequest) ToBytes() []byte {
	callerId := StringToBytes(string(r.caller))
	ptrLen := IntToBytes(uint32(len(r.ptrs)))
	ptrs := make([]byte, 0)
	for _, ptr := range r.ptrs {
		ptrs = append(ptrs, ptr.ToBytes()...)
	}
	blockId := StringToBytes(string(r.blockId))
	reqId := StringToBytes(string(r.reqId))
	return AddType(ReadRequestType, bytes.Join([][]byte{callerId, ptrLen, ptrs, blockId, reqId}, []byte{}))
}

func ToReadRequest(data []byte) *ReadRequest {
	callerId, remainder := StringFromBytes(data)
	numPtrs, remainder := IntFromBytes(remainder)
	ptrs := make([]DiskPointer, 0, numPtrs)
	for range numPtrs {
		var ptr *DiskPointer
		ptr, remainder = ToDiskPointer(remainder)
		ptrs = append(ptrs, *ptr)
	}
	blockId, remainder := StringFromBytes(remainder)
	reqId, _ := StringFromBytes(remainder)
	rq := ReadRequest{
		caller:  NodeId(callerId),
		ptrs:    ptrs,
		blockId: BlockId(blockId),
		reqId:   GetBlockId(reqId),
	}
	return &rq
}
