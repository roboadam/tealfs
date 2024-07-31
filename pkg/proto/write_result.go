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

package proto

import (
	"bytes"
)

type WriteResult struct {
	Ok      bool
	Message string
}

func (r *WriteResult) Equal(p Payload) bool {
	if o, ok := p.(*WriteResult); ok {
		if r.Ok != o.Ok {
			return false
		}
		if r.Message != o.Message {
			return false
		}
		return true
	}
	return false
}

func (r *WriteResult) ToBytes() []byte {
	ok := BoolToBytes(r.Ok)
	message := StringToBytes(r.Message)
	return bytes.Join([][]byte{ok, message}, []byte{})
}

func ToWriteResult(data []byte) *WriteResult {
	ok, remainder := BoolFromBytes(data)
	message, _ := StringFromBytes(remainder)
	return &WriteResult{
		Ok:      ok,
		Message: message,
	}
}
