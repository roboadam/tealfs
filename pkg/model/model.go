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
	"hash"

	"github.com/google/uuid"
)

type ConnectionStatus struct {
	Type          ConnectedStatus
	RemoteAddress string
	Msg           string
	Id            ConnId
}
type ConnectedStatus int

const (
	Connected ConnectedStatus = iota
	NotConnected
)

type ConnId int32

type MgrConnsConnectTo struct {
	Address string
}

type MgrConnsSend struct {
	ConnId  ConnId
	Payload Payload
}

func (m *MgrConnsSend) Equal(o *MgrConnsSend) bool {
	if m.ConnId != o.ConnId {
		return false
	}

	return m.Payload.Equal(o.Payload)
}

type ConnsMgrReceive struct {
	ConnId  ConnId
	Payload Payload
}

type MgrDiskSave struct {
	Hash hash.Hash
	Data []byte
}

type UiMgrConnectTo struct {
	Address string
}

type NodeId string

func NewNodeId() NodeId {
	idValue := uuid.New()
	return NodeId(idValue.String())
}
