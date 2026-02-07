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
	Ok      bool
	Message string
	Caller  NodeId
	Ptr     DiskPointer
	ReqId   PutBlockId
}

func NewWriteResultOk(
	ptr DiskPointer,
	caller NodeId,
	reqId PutBlockId,
) WriteResult {
	return WriteResult{
		Ok:     true,
		Caller: caller,
		Ptr:    ptr,
		ReqId:  reqId,
	}
}

func NewWriteResultErr(
	message string,
	caller NodeId,
	reqId PutBlockId,
) WriteResult {
	return WriteResult{
		Ok:      false,
		Message: message,
		Caller:  caller,
		ReqId:   reqId,
	}
}
