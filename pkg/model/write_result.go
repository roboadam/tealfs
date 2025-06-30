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

func (r *WriteResult) Ok() bool          { return r.ok }
func (r *WriteResult) Message() string   { return r.message }
func (r *WriteResult) Caller() NodeId    { return r.caller }
func (r *WriteResult) Ptr() DiskPointer  { return r.ptr }
func (r *WriteResult) ReqId() PutBlockId { return r.reqId }
