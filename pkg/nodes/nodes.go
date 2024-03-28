package nodes

type Nodes struct {
	nodes map[Id]Node
}

func New() Nodes {
	return Nodes{nodes: make(map[Id]Node, 3)}
}

func (n *Nodes) AddOrUpdate(node Node) {
	n.nodes[node.Id] = node
}

type Node struct {
	Id Id
}

type Address string
