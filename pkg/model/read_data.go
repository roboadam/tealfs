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
