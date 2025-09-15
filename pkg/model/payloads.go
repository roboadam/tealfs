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

type PayloadType uint16

const (
	NoOpType                  PayloadType = 0
	IAmType                   PayloadType = 1
	SyncType                  PayloadType = 2
	WriteRequestType          PayloadType = 3
	WriteResultType           PayloadType = 4
	ReadRequestType           PayloadType = 5
	ReadResultType            PayloadType = 6
	FileBroadcastType         PayloadType = 7
	AddDiskRequestType        PayloadType = 8
	SaveToDiskReq             PayloadType = 9
	SaveToDiskResp            PayloadType = 10
	GetFromDiskReq            PayloadType = 11
	GetFromDiskResp           PayloadType = 12
	Election                  PayloadType = 13
	Alive                     PayloadType = 14
	Victory                   PayloadType = 15
	BalanceReqType            PayloadType = 16
	OnDiskBlockIdListType     PayloadType = 17
	DeleteBlockId             PayloadType = 18
	StoreItCmdType            PayloadType = 19
	StoreItRespType           PayloadType = 20
	FilesystemBlockIdListType PayloadType = 21
)

type Payload interface {
	Type() PayloadType
}
