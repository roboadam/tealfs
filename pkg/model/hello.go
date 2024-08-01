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

type IAm struct {
	NodeId Id
}

func (h *IAm) ToBytes() []byte {
	nodeId := StringToBytes(string(h.NodeId))
	return AddType(IAmType, nodeId)
}

func (h *IAm) Equal(p Payload) bool {
	if h2, ok := p.(*IAm); ok {
		return h2.NodeId == h.NodeId
	}
	return false
}

func ToHello(data []byte) *IAm {
	rawId, _ := StringFromBytes(data)
	return &IAm{
		NodeId: Id(rawId),
	}
}
