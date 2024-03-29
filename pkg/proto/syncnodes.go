package proto

import (
	"tealfs/pkg/nodes"
	"tealfs/pkg/set"
)

type SyncNodes struct {
	Nodes set.Set[struct {
		Node    nodes.Node
		Address string
	}]
}

func (s *SyncNodes) ToBytes() []byte {
	result := make([]byte, 0)
	for _, n := range s.Nodes.GetValues() {
		result = append(result, nodeToBytes(n.Node)...)
		result = append(result, StringToBytes(n.Address)...)
	}
	return AddType(SyncType, result)
}

func nodeToBytes(node nodes.Node) []byte {
	return StringToBytes(string(node.Id))
}

func (s *SyncNodes) GetIds() set.Set[nodes.Id] {
	result := set.NewSet[nodes.Id]()
	for _, n := range s.Nodes.GetValues() {
		result.Add(n.Node.Id)
	}
	return result
}

func (s *SyncNodes) NodeForId(id nodes.Id) (nodes.Node, bool) {
	for _, n := range s.Nodes.GetValues() {
		if n.Node.Id == id {
			return n.Node, true
		}
	}
	return nodes.Node{}, false
}

func ToSyncNodes(data []byte) *SyncNodes {
	remainder := data
	result := set.NewSet[struct {
		Node    nodes.Node
		Address string
	}]()
	for {
		var n nodes.Node
		var address string
		n, remainder = toNode(remainder)
		address, remainder = StringFromBytes(remainder)
		result.Add(struct {
			Node    nodes.Node
			Address string
		}{Node: n, Address: address})
		if len(remainder) <= 0 {
			return &SyncNodes{Nodes: result}
		}
	}
}

func toNode(data []byte) (nodes.Node, []byte) {
	var id string
	remainder := data
	id, remainder = StringFromBytes(remainder)
	return nodes.Node{Id: nodes.Id(id)}, remainder
}
