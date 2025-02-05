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

import "bytes"

type IAm struct {
	NodeId    NodeId
	Address   string
	FreeBytes uint32
}

func (h *IAm) ToBytes() []byte {
	nodeId := StringToBytes(string(h.NodeId))
	address := StringToBytes(h.Address)
	freeByes := IntToBytes(h.FreeBytes)
	return AddType(IAmType, bytes.Join([][]byte{nodeId, address, freeByes}, []byte{}))
}

func (h *IAm) Equal(p Payload) bool {
	if h2, ok := p.(*IAm); ok {
		return h2.NodeId == h.NodeId && h2.Address == h.Address && h2.FreeBytes == h.FreeBytes
	}
	return false
}

func ToHello(data []byte) *IAm {
	rawId, remainder := StringFromBytes(data)
	rawAddress, remainder := StringFromBytes(remainder)
	rawFreeBytes, _ := IntFromBytes(remainder)
	return &IAm{
		NodeId:    NodeId(rawId),
		Address:   rawAddress,
		FreeBytes: rawFreeBytes,
	}
}
