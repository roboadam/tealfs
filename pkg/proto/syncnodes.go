package proto

import (
	"tealfs/pkg/model/node"
	"tealfs/pkg/util"
)

type SyncNodes struct {
	Nodes util.Set[node.Node]
}

func (s *SyncNodes) ToBytes() []byte {
	result := make([]byte, 0)
	for _, n := range s.Nodes.GetValues() {
		result = append(result, nodeToBytes(n)...)
	}
	return AddType(SyncType, result)
}

func nodeToBytes(node node.Node) []byte {
	result := make([]byte, 0)
	id := StringToBytes(node.Id.String())
	address := StringToBytes(node.Address.Value)
	result = append(result, id...)
	result = append(result, address...)
	return result
}

func (s *SyncNodes) GetIds() util.Set[node.Id] {
	result := util.NewSet[node.Id]()
	for _, n := range s.Nodes.GetValues() {
		result.Add(n.Id)
	}
	return result
}

func (s *SyncNodes) NodeForId(id node.Id) (node.Node, bool) {
	for _, n := range s.Nodes.GetValues() {
		if n.Id == id {
			return n, true
		}
	}
	return node.Node{}, false
}

func ToSyncNodes(data []byte) *SyncNodes {
	remainder := data
	result := util.NewSet[node.Node]()
	for {
		var n node.Node
		n, remainder = toNode(remainder)
		result.Add(n)
		if len(remainder) <= 0 {
			return &SyncNodes{Nodes: result}
		}
	}
}

func toNode(data []byte) (node.Node, []byte) {
	var id, address string
	remainder := data
	id, remainder = StringFromBytes(remainder)
	address, remainder = StringFromBytes(remainder)
	return node.Node{Id: node.IdFromRaw(id), Address: node.NewAddress(address)}, remainder
}
