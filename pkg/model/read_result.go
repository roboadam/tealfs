// Copyright (C) 2026 Adam Hess
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
	Ok      bool
	Message string
	Caller  NodeId
	Ptrs    []DiskPointer
	Data    RawData
	BlockId BlockId
	ReqId   GetBlockId
}

func (r *ReadResult) Type() PayloadType {
	return ReadResultType
}

func NewReadResultOk(
	caller NodeId,
	ptrs []DiskPointer,
	data RawData,
	reqId GetBlockId,
	blockId BlockId,
) ReadResult {
	return ReadResult{
		Ok:      true,
		Message: "",
		Caller:  caller,
		Ptrs:    ptrs,
		Data:    data,
		ReqId:   reqId,
		BlockId: blockId,
	}
}

func NewReadResultErr(
	message string,
	caller NodeId,
	reqId GetBlockId,
	blockId BlockId,
) ReadResult {
	return ReadResult{
		Ok:      false,
		Message: message,
		Caller:  caller,
		ReqId:   reqId,
		BlockId: blockId,
	}
}
