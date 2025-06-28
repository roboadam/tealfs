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

type WriteRequest struct {
	caller NodeId
	data   RawData
	reqId  PutBlockId
}

func NewWriteRequest(
	caller NodeId,
	data RawData,
	reqId PutBlockId,
) WriteRequest {
	return WriteRequest{
		caller: caller,
		data:   data,
		reqId:  reqId,
	}
}

func (r *WriteRequest) Caller() NodeId    { return r.caller }
func (r *WriteRequest) Data() RawData     { return r.data }
func (r *WriteRequest) ReqId() PutBlockId { return r.reqId }

func (r *WriteRequest) ToBytes() []byte {
	caller := StringToBytes(string(r.caller))
	rawData := r.data.ToBytes()
	putblockId := StringToBytes(string(r.reqId))
	payload := bytes.Join([][]byte{caller, rawData, putblockId}, []byte{})
	return AddType(WriteRequestType, payload)
}

func ToWriteRequest(raw []byte) *WriteRequest {
	caller, remainder := StringFromBytes(raw)
	rawData, remainder := ToRawData(remainder)
	putBlockId, _ := StringFromBytes(remainder)
	return &WriteRequest{
		caller: NodeId(caller),
		data:   *rawData,
		reqId:  PutBlockId(putBlockId),
	}
}
