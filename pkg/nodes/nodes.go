package nodes

type Nodes struct {
	nodes map[Id]NodeNew
}

func New() Nodes {
	return Nodes{nodes: make(map[Id]NodeNew, 3)}
}

func (n *Nodes) AddOrUpdate(node NodeNew) {
	n.nodes[node.Id] = node
}

type NodeNew struct {
	Id Id
}

type Address string
