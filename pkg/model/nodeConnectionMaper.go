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
	"encoding/json"
	"sync"
	"tealfs/pkg/set"
)

type NodeConnectionMapper struct {
	addresses      set.Set[string]
	addressConnMap set.Bimap[string, ConnId]
	connNodeMap    set.Bimap[ConnId, NodeId]
	addressNodeMap set.Bimap[string, NodeId]
	mux            sync.RWMutex
}

type NodeConnectionMapperExport struct {
	Addresses      []string
	AddressConnMap map[string]ConnId
	ConnNodeMap    map[ConnId]NodeId
	AddressNodeMap map[string]NodeId
}

func (n *NodeConnectionMapper) AddressesWithoutConnections() set.Set[string] {
	n.mux.RLock()
	defer n.mux.RUnlock()
	result := set.NewSet[string]()
	for _, address := range n.addresses.GetValues() {
		if _, ok := n.addressConnMap.Get1(address); !ok {
			result.Add(address)
		}
	}
	return result
}

func (n *NodeConnectionMapper) Connections() set.Set[ConnId] {
	n.mux.RLock()
	defer n.mux.RUnlock()
	return n.connections()
}

func (n *NodeConnectionMapper) connections() set.Set[ConnId] {
	result := set.NewSet[ConnId]()
	for _, address := range n.addresses.GetValues() {
		if conn, ok := n.addressConnMap.Get1(address); ok {
			result.Add(conn)
		}
	}
	return result
}

func (n *NodeConnectionMapper) ConnForNode(node NodeId) (ConnId, bool) {
	n.mux.RLock()
	defer n.mux.RUnlock()
	return n.connNodeMap.Get2(node)
}

func (n *NodeConnectionMapper) NodeForConn(connId ConnId) (NodeId, bool) {
	n.mux.RLock()
	defer n.mux.RUnlock()
	node, ok := n.connNodeMap.Get1(connId)
	return node, ok
}

func (n *NodeConnectionMapper) AddressForConn(connId ConnId) (string, bool) {
	n.mux.RLock()
	defer n.mux.RUnlock()
	return n.addressConnMap.Get2(connId)
}

func (n *NodeConnectionMapper) RemoveConn(connId ConnId) {
	n.mux.Lock()
	defer n.mux.Unlock()
	n.removeConn(connId)
}

func (n *NodeConnectionMapper) removeConn(connId ConnId) {
	n.addressConnMap.Remove2(connId)
	n.connNodeMap.Remove1(connId)
}

func (n *NodeConnectionMapper) Marshal() ([]byte, error) {
	n.mux.RLock()
	defer n.mux.RUnlock()
	exportable := NodeConnectionMapperExport{
		Addresses:      n.addresses.ToSlice(),
		AddressConnMap: n.addressConnMap.ToMap(),
		ConnNodeMap:    n.connNodeMap.ToMap(),
		AddressNodeMap: n.addressNodeMap.ToMap(),
	}
	return json.Marshal(exportable)
}

func NodeConnectionMapperUnmarshal(data []byte) (*NodeConnectionMapper, error) {
	var exportable NodeConnectionMapperExport
	err := json.Unmarshal(data, &exportable)
	if err != nil {
		return nil, err
	}
	result := NodeConnectionMapper{
		addresses:      set.NewSetFromSlice(exportable.Addresses),
		addressConnMap: set.NewBimapFromMap(exportable.AddressConnMap),
		connNodeMap:    set.NewBimapFromMap(exportable.ConnNodeMap),
		addressNodeMap: set.NewBimapFromMap(exportable.AddressNodeMap),
		mux:            sync.RWMutex{},
	}

	return &result, nil
}

func (n *NodeConnectionMapper) AddressesAndNodes() set.Set[struct {
	Address string
	NodeId  NodeId
}] {
	n.mux.RLock()
	defer n.mux.RUnlock()
	result := set.NewSet[struct {
		Address string
		NodeId  NodeId
	}]()
	for _, values := range n.addressNodeMap.AllValues() {
		result.Add(struct {
			Address string
			NodeId  NodeId
		}{
			Address: values.K,
			NodeId:  values.J,
		})
	}
	return result
}
func (n *NodeConnectionMapper) SetAll(conn ConnId, address string, node NodeId) {
	n.mux.Lock()
	defer n.mux.Unlock()
	n.addresses.Add(address)
	n.addressConnMap.Add(address, conn)
	n.connNodeMap.Add(conn, node)
	n.addressNodeMap.Add(address, node)
}

func (n *NodeConnectionMapper) Nodes() set.Set[NodeId] {
	n.mux.RLock()
	defer n.mux.RUnlock()
	result := set.NewSet[NodeId]()
	for _, values := range n.addressNodeMap.AllValues() {
		result.Add(values.J)
	}
	return result
}

func (n *NodeConnectionMapper) UnsetConnections() {
	n.mux.Lock()
	defer n.mux.Unlock()
	connections := n.connections()
	for _, conn := range connections.GetValues() {
		n.removeConn(conn)
	}
}

func NewNodeConnectionMapper() *NodeConnectionMapper {
	return &NodeConnectionMapper{
		addresses:      set.NewSet[string](),
		addressConnMap: set.NewBimap[string, ConnId](),
		connNodeMap:    set.NewBimap[ConnId, NodeId](),
		addressNodeMap: set.NewBimap[string, NodeId](),
		mux:            sync.RWMutex{},
	}
}
