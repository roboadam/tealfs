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

type ErrorResp struct {
	msg string
}

func NewErrorResp(msg string) ErrorResp {
	return ErrorResp{msg: msg}
}

func (e *ErrorResp) Msg() string { return e.msg }

func (e *ErrorResp) ToBytes() []byte {
	msg := StringToBytes(e.msg)
	return AddType(ErrorRespType, msg)
}

func (e *ErrorResp) Equal(p Payload) bool {
	if e2, ok := p.(*ErrorResp); ok {
		return e2.msg == e.msg
	}
	return false
}

func ToErrorResp(data []byte) (ErrorResp, []byte) {
	msg, remainder := StringFromBytes(data)
	return ErrorResp{msg: msg}, remainder
}
