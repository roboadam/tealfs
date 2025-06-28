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

const (
	NoOpType         = uint8(0)
	IAmType          = uint8(1)
	SyncType         = uint8(2)
	WriteRequestType = uint8(3)
	WriteResultType  = uint8(4)
	ReadRequestType  = uint8(5)
	ReadResultType   = uint8(6)
	BroadcastType    = uint8(7)
	AddDiskRequest   = uint8(8)
)

type Payload2 any
