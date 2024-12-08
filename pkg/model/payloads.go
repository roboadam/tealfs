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

const (
	NoOpType                = uint8(0)
	IAmType                 = uint8(1)
	SyncType                = uint8(2)
	WriteRequestType        = uint8(3)
	WriteResultType         = uint8(4)
	ReadRequestType         = uint8(5)
	ReadResultType          = uint8(6)
	LockConfirmRequestType  = uint8(7)
	LockConfirmResponseType = uint8(8)
	LockCreateRequestType   = uint8(9)
)

type Payload interface {
	ToBytes() []byte
	Equal(Payload) bool
}

func ToPayload(data []byte) Payload {
	switch payloadType(data) {
	case IAmType:
		return ToHello(payloadData(data))
	case SyncType:
		return ToSyncNodes(payloadData(data))
	case WriteRequestType:
		return ToWriteRequest(payloadData(data))
	case WriteResultType:
		return ToWriteResult(payloadData(data))
	case ReadRequestType:
		return ToReadRequest(payloadData(data))
	case ReadResultType:
		return ToReadResult(payloadData(data))
	case LockConfirmRequestType:
		return ToLockConfirmRequest(payloadData(data))
	case LockConfirmResponseType:
		return ToLockConfirmResponse(payloadData(data))
	case LockCreateRequestType:
		return ToLockCreateRequest(payloadData(data))
	default:
		return ToNoOp(payloadData(data))
	}
}

func payloadData(data []byte) []byte {
	if len(data) > 0 {
		return data[1:]
	} else {
		return data
	}
}

func payloadType(data []byte) byte {
	if len(data) <= 0 {
		return NoOpType
	}
	return data[0]
}
