package mgr

import (
	"tealfs/pkg/nodes"
)

type NodeConnMap struct {
	nodeToConn map[nodes.Id]ConnId
	connToNode map[ConnId]nodes.Id
}

func (n *NodeConnMap) Add(node nodes.Id, conn ConnId) {
	n.nodeToConn[node] = conn
	n.connToNode[conn] = node
}

func (n *NodeConnMap) Node(conn ConnId) (nodes.Id, bool) {
	result, ok := n.connToNode[conn]
	return result, ok
}
func (n *NodeConnMap) Conn(node nodes.Id) (ConnId, bool) {
	result, ok := n.nodeToConn[node]
	return result, ok
}
