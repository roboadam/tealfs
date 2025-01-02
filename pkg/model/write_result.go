// Copyright (C) 2024 Adam Hess
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
	Ok       bool
	Message  string
	Caller   NodeId
	BlockKey BlockKey
}

func (r *WriteResult) Equal(p Payload) bool {
	if o, ok := p.(*WriteResult); ok {
		if r.Ok != o.Ok {
			return false
		}
		if r.Message != o.Message {
			return false
		}
		if r.Caller != o.Caller {
			return false
		}
		if !r.BlockKey.Equals(&o.BlockKey) {
			return false
		}
		return true
	}
	return false
}

func (r *WriteResult) ToBytes() []byte {
	ok := BoolToBytes(r.Ok)
	message := StringToBytes(r.Message)
	caller := StringToBytes(string(r.Caller))
	blockKey := r.BlockKey.ToBytes()

	payload := bytes.Join([][]byte{ok, message, caller, blockKey}, []byte{})
	return AddType(WriteResultType, payload)
}

func ToWriteResult(data []byte) *WriteResult {
	ok, remainder := BoolFromBytes(data)
	message, remainder := StringFromBytes(remainder)
	caller, remainder := StringFromBytes(remainder)
	blockKey, _ := ToBlockKey(remainder)
	return &WriteResult{
		Ok:       ok,
		Message:  message,
		Caller:   NodeId(caller),
		BlockKey: *blockKey,
	}
}
