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
	"hash"

	"github.com/google/uuid"
)

type UiConnectionStatus struct {
	Type          ConnectedStatus
	RemoteAddress string
	Msg           string
	Id            NodeId
}
type UiDiskStatus struct {
	Localness     Localness
	Availableness DiskAvailableness
	Node          NodeId
	Id            DiskId
	Path          string
}
type Localness int

const (
	Local Localness = iota
	Remote
)

type DiskAvailableness int

const (
	Available DiskAvailableness = iota
	Unavailable
	Unknown
)

type NetConnectionStatus struct {
	Type ConnectedStatus
	Msg  string
	Id   ConnId
}
type ConnectedStatus int

const (
	Connected ConnectedStatus = iota
	NotConnected
)

type ConnId int32

type ConnectToNodeReq struct {
	Address string
}

type MgrConnsSend struct {
	ConnId  ConnId
	Payload Payload
}

type ConnsMgrReceive struct {
	ConnId  ConnId
	Payload Payload
}

type MgrDiskSave struct {
	Hash hash.Hash
	Data []byte
}

type AddDiskReq struct {
	Path      string
	Node      NodeId
	FreeBytes int
}

func (a *AddDiskReq) Type() PayloadType {
	return AddDiskRequestType
}

func NewAddDiskReq(
	path string,
	node NodeId,
	freeBytes int,
) AddDiskReq {
	return AddDiskReq{
		Path:      path,
		Node:      node,
		FreeBytes: freeBytes,
	}
}

type NodeId string

func NewNodeId() NodeId {
	idValue := uuid.New()
	return NodeId(idValue.String())
}

type DiskId string

type DiskIdPath struct {
	Id   DiskId
	Path string
	Node NodeId
}
