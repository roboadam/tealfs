package nodes

import "tealfs/pkg/set"

type Nodes struct {
	nodes map[Id]Node
}

func New() Nodes {
	return Nodes{nodes: make(map[Id]Node, 3)}
}

func (n *Nodes) AddOrUpdate(node Node) {
	n.nodes[node.Id] = node
}

func (n *Nodes) ToSet() set.Set[Node] {
	result := set.NewSet[Node]()
	for _, node := range n.nodes {
		result.Add(node)
	}
	return result
}

type Node struct {
	Id Id
}

type Address string
