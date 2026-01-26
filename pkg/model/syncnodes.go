// Copyright (C) 2026 Adam Hess
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
	"tealfs/pkg/set"
)

type SyncNodes struct {
	Nodes set.Set[struct {
		Node    NodeId
		Address string
	}]
}

func (s *SyncNodes) Type() PayloadType {
	return SyncType
}

func NewSyncNodes() SyncNodes {
	return SyncNodes{
		Nodes: set.NewSet[struct {
			Node    NodeId
			Address string
		}](),
	}
}

func (s *SyncNodes) GetNodes() set.Set[NodeId] {
	result := set.NewSet[NodeId]()
	for _, n := range s.Nodes.GetValues() {
		result.Add(n.Node)
	}
	return result
}

func (s *SyncNodes) AddressForNode(id NodeId) string {
	for _, val := range s.Nodes.GetValues() {
		if val.Node == id {
			return val.Address
		}
	}
	return ""
}
