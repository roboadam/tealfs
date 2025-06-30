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
