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

type connRecord struct {
	address *string
	nodeId  *NodeId
}

type addressRecord struct {
	nodeId *NodeId
	connId *ConnId
}

type nodeRecord struct {
	address *string
	connId  *ConnId
}

type NodeConnectionMapper struct {
	connRecords    map[ConnId]connRecord
	addressRecords map[string]addressRecord
	nodeRecords    map[NodeId]nodeRecord
}

func (n *NodeConnectionMapper) AddressesWithoutConnections() []string {
	result := []string{}
	for address := range n.addressRecords {
		if n.addressRecords[address].connId == nil {
			result = append(result, address)
		}
	}
	return result
}

func (n *NodeConnectionMapper) Connections() []ConnId {
	result := []ConnId{}
	for conn := range n.connRecords {
		result = append(result, conn)
	}
	return result
}

func (n *NodeConnectionMapper) ConnForNode(node NodeId) (ConnId, bool) {
	if record, ok := n.nodeRecords[node]; ok {
		return *record.connId, true
	}
	return -1, false
}

func (n *NodeConnectionMapper) AddressesAndNodes() []struct {
	Address string
	NodeId  NodeId
} {
	result := []struct {
		Address string
		NodeId  NodeId
	}{}
	for address := range n.addressRecords {
		if n.addressRecords[address].nodeId != nil {
			result = append(result, struct {
				Address string
				NodeId  NodeId
			}{
				Address: address,
				NodeId:  *n.addressRecords[address].nodeId,
			})
		}
	}
	return result
}

func NewNodeConnectionMapper() *NodeConnectionMapper {
	return &NodeConnectionMapper{
		connRecords:    make(map[ConnId]connRecord),
		addressRecords: make(map[string]addressRecord),
		nodeRecords:    make(map[NodeId]nodeRecord),
	}
}
