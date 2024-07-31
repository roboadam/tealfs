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

package mgr

import (
	"tealfs/pkg/model"
	"tealfs/pkg/nodes"
)

type NodeConnMap struct {
	nodeToConn map[nodes.Id]model.ConnId
	connToNode map[model.ConnId]nodes.Id
}

func (n *NodeConnMap) Add(node nodes.Id, conn model.ConnId) {
	n.nodeToConn[node] = conn
	n.connToNode[conn] = node
}

func (n *NodeConnMap) Node(conn model.ConnId) (nodes.Id, bool) {
	result, ok := n.connToNode[conn]
	return result, ok
}
func (n *NodeConnMap) Conn(node nodes.Id) (model.ConnId, bool) {
	result, ok := n.nodeToConn[node]
	return result, ok
}
