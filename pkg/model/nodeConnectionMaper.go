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

import "tealfs/pkg/set"

type NodeConnectionMapper struct {
	addresses      set.Set[string]
	addressConnMap set.Bimap[string, ConnId]
	connNodeMap    set.Bimap[ConnId, *NodeId]
	addressNodeMap set.Bimap[string, *NodeId]
}

func (n *NodeConnectionMapper) AddressesWithoutConnections() []string {
	result := []string{}
	for _, address := range n.addresses.GetValues() {
		if _, ok := n.addressConnMap.Get1(address); !ok {
			result = append(result, address)
		}
	}
	return result
}

func (n *NodeConnectionMapper) Connections() []ConnId {
	result := []ConnId{}
	for _, address := range n.addresses.GetValues() {
		if conn, ok := n.addressConnMap.Get1(address); ok {
			result = append(result, conn)
		}
	}
	return result
}

func (n *NodeConnectionMapper) ConnForNode(node NodeId) (ConnId, bool) {
	return n.connNodeMap.Get2(&node)
}

func (n *NodeConnectionMapper) AddressesAndNodes() []struct {
	Address string
	NodeId  NodeId
} {
	result := []struct {
		Address string
		NodeId  NodeId
	}{}
	for _, values := range n.addressNodeMap.AllValues() {
		result = append(result, struct {
			Address string
			NodeId  NodeId
		}{
			Address: values.K,
			NodeId:  *values.J,
		})
	}
	return result
}
func (n *NodeConnectionMapper) SetAll(conn ConnId, address string, node NodeId) {
	n.addresses.Add(address)
	n.addressConnMap.Add(address, conn)
	n.connNodeMap.Add(conn, &node)
	n.addressNodeMap.Add(address, &node)
}

func NewNodeConnectionMapper() *NodeConnectionMapper {
	return &NodeConnectionMapper{
		addresses:      set.NewSet[string](),
		addressConnMap: set.NewBimap[string, ConnId](),
		connNodeMap:    set.NewBimap[ConnId, *NodeId](),
		addressNodeMap: set.NewBimap[string, *NodeId](),
	}
}
