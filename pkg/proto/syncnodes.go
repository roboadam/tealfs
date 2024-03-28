package proto

import (
	"tealfs/pkg/nodes"
	"tealfs/pkg/set"
)

type SyncNodes struct {
	Nodes set.Set[nodes.NodeNew]
}

func (s *SyncNodes) ToBytes() []byte {
	result := make([]byte, 0)
	for _, n := range s.Nodes.GetValues() {
		result = append(result, nodeToBytes(n)...)
	}
	return AddType(SyncType, result)
}

func nodeToBytes(node nodes.NodeNew) []byte {
	return StringToBytes(string(node.Id))
}

func (s *SyncNodes) GetIds() set.Set[nodes.Id] {
	result := set.NewSet[nodes.Id]()
	for _, n := range s.Nodes.GetValues() {
		result.Add(n.Id)
	}
	return result
}

func (s *SyncNodes) NodeForId(id nodes.Id) (nodes.NodeNew, bool) {
	for _, n := range s.Nodes.GetValues() {
		if n.Id == id {
			return n, true
		}
	}
	return nodes.NodeNew{}, false
}

func ToSyncNodes(data []byte) *SyncNodes {
	remainder := data
	result := set.NewSet[nodes.NodeNew]()
	for {
		var n nodes.NodeNew
		n, remainder = toNode(remainder)
		result.Add(n)
		if len(remainder) <= 0 {
			return &SyncNodes{Nodes: result}
		}
	}
}

func toNode(data []byte) (nodes.NodeNew, []byte) {
	var id string
	remainder := data
	id, remainder = StringFromBytes(remainder)
	return nodes.NodeNew{Id: nodes.Id(id)}, remainder
}
