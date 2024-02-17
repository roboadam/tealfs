package mgr

type NodesNew struct {
	nodes map[NodeNewId]NodeNew
}

func (n *NodesNew) AddOrUpdate(node NodeNew) {
	n.nodes[node.id] = node
}

type NodeNew struct {
	id      NodeNewId
	address NodeNewAddress
}

type NodeNewAddress string

type NodeNewId int32
