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

type WriteResult struct {
	ok      bool
	message string
	caller  NodeId
	ptr     DiskPointer
	reqId   PutBlockId
}

func NewWriteResultOk(
	ptr DiskPointer,
	caller NodeId,
	reqId PutBlockId,
) WriteResult {
	return WriteResult{
		ok:     true,
		caller: caller,
		ptr:    ptr,
		reqId:  reqId,
	}
}

func NewWriteResultErr(
	message string,
	caller NodeId,
	reqId PutBlockId,
) WriteResult {
	return WriteResult{
		ok:      false,
		message: message,
		caller:  caller,
		reqId:   reqId,
	}
}

func NewWriteResultSuccess(
	caller NodeId,
	ptr DiskPointer,
	reqId PutBlockId,
) WriteResult {
	return WriteResult{
		ok:     true,
		caller: caller,
		ptr:    ptr,
		reqId:  reqId,
	}
}

func (r *WriteResult) Equal(p Payload) bool {
	if o, ok := p.(*WriteResult); ok {
		if r.ok != o.ok {
			return false
		}
		if r.message != o.message {
			return false
		}
		if r.caller != o.caller {
			return false
		}
		if !r.ptr.Equals(&o.ptr) {
			return false
		}
		if r.reqId != o.reqId {
			return false
		}
		return true
	}
	return false
}

func (r *WriteResult) Ok() bool          { return r.ok }
func (r *WriteResult) Message() string   { return r.message }
func (r *WriteResult) Caller() NodeId    { return r.caller }
func (r *WriteResult) Ptr() DiskPointer  { return r.ptr }
func (r *WriteResult) ReqId() PutBlockId { return r.reqId }

func (r *WriteResult) ToBytes() []byte {
	ok := BoolToBytes(r.ok)
	message := StringToBytes(r.message)
	caller := StringToBytes(string(r.caller))
	ptr := r.ptr.ToBytes()
	reqId := StringToBytes(string(r.reqId))

	payload := bytes.Join([][]byte{ok, message, caller, ptr, reqId}, []byte{})
	return AddType(WriteResultType, payload)
}

func ToWriteResult(data []byte) *WriteResult {
	ok, remainder := BoolFromBytes(data)
	message, remainder := StringFromBytes(remainder)
	caller, remainder := StringFromBytes(remainder)
	ptr, remainder := ToDiskPointer(remainder)
	reqId, _ := StringFromBytes(remainder)
	return &WriteResult{
		ok:      ok,
		message: message,
		caller:  NodeId(caller),
		ptr:     *ptr,
		reqId:   PutBlockId(reqId),
	}
}
