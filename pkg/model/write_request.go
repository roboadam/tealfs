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
	caller     NodeId
	data       RawData
	putBlockId PutBlockId
}

func NewWriteRequest(
	caller NodeId,
	data RawData,
	putBlockId PutBlockId,
) WriteRequest {
	return WriteRequest{
		caller:     caller,
		data:       data,
		putBlockId: putBlockId,
	}
}

func (r *WriteRequest) Caller() NodeId         { return r.caller }
func (r *WriteRequest) Data() RawData          { return r.data }
func (r *WriteRequest) PutBlockId() PutBlockId { return r.putBlockId }

func (r *WriteRequest) Equal(p Payload) bool {
	if o, ok := p.(*WriteRequest); ok {
		if r.caller != o.caller {
			return false
		}
		if r.putBlockId != o.putBlockId {
			return false
		}
		return r.data.Equals(&o.data)
	}
	return false
}

func (r *WriteRequest) ToBytes() []byte {
	caller := StringToBytes(string(r.caller))
	rawData := r.data.ToBytes()
	putblockId := StringToBytes(string(r.putBlockId))
	payload := bytes.Join([][]byte{caller, rawData, putblockId}, []byte{})
	return AddType(WriteRequestType, payload)
}

func ToWriteRequest(raw []byte) *WriteRequest {
	caller, remainder := StringFromBytes(raw)
	rawData, remainder := ToRawData(remainder)
	putBlockId, _ := StringFromBytes(remainder)
	return &WriteRequest{
		caller:     NodeId(caller),
		data:       *rawData,
		putBlockId: PutBlockId(putBlockId),
	}
}
