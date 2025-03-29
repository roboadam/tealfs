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
	"bytes"
)

type IAm struct {
	nodeId    NodeId
	disks     []DiskIdPath
	address   string
	freeBytes uint32
}

func NewIam(nodeId NodeId, disks []DiskIdPath, address string, freeBytes uint32) IAm {
	return IAm{
		nodeId:    nodeId,
		disks:     disks,
		address:   address,
		freeBytes: freeBytes,
	}
}

func (i *IAm) Node() NodeId        { return i.nodeId }
func (i *IAm) Disks() []DiskIdPath { return i.disks }
func (i *IAm) Address() string     { return i.address }
func (i *IAm) FreeBytes() uint32   { return i.freeBytes }

func (h *IAm) ToBytes() []byte {
	nodeId := StringToBytes(string(h.nodeId))
	disksLen := IntToBytes(uint32(len(h.disks)))
	diskBytes := []byte{}
	for _, disk := range h.disks {
		diskBytes = append(diskBytes, disk.ToBytes()...)
	}
	address := StringToBytes(h.address)
	freeByes := IntToBytes(h.freeBytes)
	return AddType(IAmType, bytes.Join([][]byte{nodeId, disksLen, diskBytes, address, freeByes}, []byte{}))
}

func (d DiskIdPath) ToBytes() []byte {
	id := StringToBytes(string(d.Id))
	path := StringToBytes(d.Path)
	node := StringToBytes(string(d.Node))
	return bytes.Join([][]byte{id, path, node}, []byte{})
}

func (h *IAm) Equal(p Payload) bool {
	if h2, ok := p.(*IAm); ok {
		if h2.nodeId != h.nodeId {
			return false
		}
		if len(h2.disks) != len(h.disks) {
			return false
		}
		for i, disk := range h2.disks {
			if disk != h.disks[i] {
				return false
			}
		}
		if h2.address != h.address {
			return false
		}
		if h2.freeBytes != h.freeBytes {
			return false
		}
		return true
	}
	return false
}

func ToDiskIdPath(data []byte) (DiskIdPath, []byte) {
	id, remainder := StringFromBytes(data)
	path, remainder := StringFromBytes(remainder)
	nodeId, remainder := StringFromBytes(remainder)
	return DiskIdPath{Id: DiskId(id), Path: path, Node: NodeId(nodeId)}, remainder
}

func ToHello(data []byte) *IAm {
	rawId, remainder := StringFromBytes(data)
	diskLen, remainder := IntFromBytes(remainder)
	disks := []DiskIdPath{}
	for range diskLen {
		var d DiskIdPath
		d, remainder = ToDiskIdPath(remainder)
		disks = append(disks, d)
	}
	rawAddress, remainder := StringFromBytes(remainder)
	rawFreeBytes, _ := IntFromBytes(remainder)
	return &IAm{
		nodeId:    NodeId(rawId),
		disks:     disks,
		address:   rawAddress,
		freeBytes: rawFreeBytes,
	}
}
