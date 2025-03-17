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
	disks     []DiskId
	address   string
	freeBytes uint32
}

func NewIam(nodeId NodeId, disks []DiskId, address string, freeBytes uint32) IAm {
	return IAm{
		nodeId:    nodeId,
		disks:     disks,
		address:   address,
		freeBytes: freeBytes,
	}
}

func (i *IAm) Node() NodeId      { return i.nodeId }
func (i *IAm) Disks() []DiskId   { return i.disks }
func (i *IAm) Address() string   { return i.address }
func (i *IAm) FreeBytes() uint32 { return i.freeBytes }

func (h *IAm) ToBytes() []byte {
	nodeId := StringToBytes(string(h.nodeId))
	disksLen := IntToBytes(uint32(len(h.disks)))
	diskBytes := []byte{}
	for _, disk := range h.disks {
		diskBytes = append(diskBytes, StringToBytes(string(disk))...)
	}
	address := StringToBytes(h.address)
	freeByes := IntToBytes(h.freeBytes)
	return AddType(IAmType, bytes.Join([][]byte{nodeId, disksLen, diskBytes, address, freeByes}, []byte{}))
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

func ToHello(data []byte) *IAm {
	rawId, remainder := StringFromBytes(data)
	diskLen, remainder := IntFromBytes(remainder)
	disks := []DiskId{}
	for range diskLen {
		var diskStr string
		diskStr, remainder = StringFromBytes(remainder)
		disk := DiskId(diskStr)
		disks = append(disks, disk)
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
