package nodes

type Nodes struct {
	nodes map[Id]NodeNew
}

func New() Nodes {
	return Nodes{nodes: make(map[Id]NodeNew, 3)}
}

func (n *Nodes) AddOrUpdate(node NodeNew) {
	n.nodes[node.id] = node
}

type NodeNew struct {
	id      Id
	address Address
}

type Address string
