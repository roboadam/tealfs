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
	for _, node := range s.Nodes.GetValues() {
		id := StringToBytes(node.Id.String())
		address := StringToBytes(node.Address.Value)
		result = append(result, id...)
		result = append(result, address...)
	}
	return AddType(SyncType, result)
}

func (s *SyncNodes) GetIds() util.Set[node.Id] {
	result := util.NewSet[node.Id]()
	for _, node := range s.Nodes.GetValues() {
		result.Add(node.Id)
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
		var id, address string
		id, remainder = StringFromBytes(remainder)
		address, remainder = StringFromBytes(remainder)
		node := node.Node{Id: node.IdFromRaw(id), Address: node.NewAddress(address)}
		result.Add(node)
		if len(remainder) <= 0 {
			return &SyncNodes{Nodes: result}
		}
	}
}
