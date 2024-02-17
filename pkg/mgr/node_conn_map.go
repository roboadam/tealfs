package mgr

import (
	"tealfs/pkg/nodes"
)

type NodeConnMap struct {
	nodeToConn map[nodes.Id]ConnNewId
	connToNode map[ConnNewId]NodeNewId
}

func (n *NodeConnMap) Add(node NodeNewId, conn ConnNewId) {
	n.nodeToConn[node] = conn
	n.connToNode[conn] = node
}

func (n *NodeConnMap) Node(conn ConnNewId) (NodeNewId, bool) {
	result, ok := n.connToNode[conn]
	return result, ok
}
func (n *NodeConnMap) Conn(node NodeNewId) (ConnNewId, bool) {
	result, ok := n.nodeToConn[node]
	return result, ok
}
