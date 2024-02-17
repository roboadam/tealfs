package nodes

type Nodes struct {
	nodes map[Id]NodeNew
}

func (n *Nodes) AddOrUpdate(node NodeNew) {
	n.nodes[node.id] = node
}

type NodeNew struct {
	id      Id
	address Address
}

type Address string
