package mgr

import (
	"tealfs/pkg/nodes"
)

type NodeConnMap struct {
	nodeToConn map[nodes.Id]ConnNewId
	connToNode map[ConnNewId]nodes.Id
}

func (n *NodeConnMap) Add(node nodes.Id, conn ConnNewId) {
	n.nodeToConn[node] = conn
	n.connToNode[conn] = node
}

func (n *NodeConnMap) Node(conn ConnNewId) (nodes.Id, bool) {
	result, ok := n.connToNode[conn]
	return result, ok
}
func (n *NodeConnMap) Conn(node nodes.Id) (ConnNewId, bool) {
	result, ok := n.nodeToConn[node]
	return result, ok
}
