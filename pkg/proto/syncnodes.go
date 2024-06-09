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

package proto

import (
	"tealfs/pkg/nodes"
	"tealfs/pkg/set"
)

type SyncNodes struct {
	Nodes set.Set[struct {
		Node    nodes.Id
		Address string
	}]
}

func NewSyncNodes() SyncNodes {
	return SyncNodes{
		Nodes: set.NewSet[struct {
			Node    nodes.Id
			Address string
		}](),
	}
}

func (s *SyncNodes) Equal(p Payload) bool {
	if s2, ok := p.(*SyncNodes); ok {
		return s.Nodes.Equal(&s2.Nodes)
	}
	return false
}

func (s *SyncNodes) ToBytes() []byte {
	result := make([]byte, 0)
	for _, n := range s.Nodes.GetValues() {
		result = append(result, nodeToBytes(n.Node)...)
		result = append(result, StringToBytes(n.Address)...)
	}
	return AddType(SyncType, result)
}

func nodeToBytes(node nodes.Id) []byte {
	return StringToBytes(string(node))
}

func (s *SyncNodes) GetNodes() set.Set[nodes.Id] {
	result := set.NewSet[nodes.Id]()
	for _, n := range s.Nodes.GetValues() {
		result.Add(n.Node)
	}
	return result
}

func (s *SyncNodes) AddressForNode(id nodes.Id) string {
	for _, val := range s.Nodes.GetValues() {
		if val.Node == id {
			return val.Address
		}
	}
	return ""
}

func ToSyncNodes(data []byte) *SyncNodes {
	remainder := data
	result := set.NewSet[struct {
		Node    nodes.Id
		Address string
	}]()
	for {
		var n nodes.Id
		var address string
		n, remainder = toNode(remainder)
		address, remainder = StringFromBytes(remainder)
		result.Add(struct {
			Node    nodes.Id
			Address string
		}{Node: n, Address: address})
		if len(remainder) <= 0 {
			return &SyncNodes{Nodes: result}
		}
	}
}

func toNode(data []byte) (nodes.Id, []byte) {
	var id string
	remainder := data
	id, remainder = StringFromBytes(remainder)
	return nodes.Id(id), remainder
}
