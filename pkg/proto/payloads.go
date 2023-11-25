package proto

import (
	"tealfs/pkg/node"
	"tealfs/pkg/util"
)

const (
	NoOpType  = uint8(0)
	HelloType = uint8(1)
	SyncType  = uint8(2)
)

type Payload interface {
	ToBytes() []byte
}

func ToPayload(data []byte) Payload {
	switch payloadType(data) {
	case HelloType:
		return ToHello(data)
	default:
		return ToNoOp(data)
	}
}

type Hello struct {
	NodeId node.Id
}

type SyncNodes struct {
	Nodes util.Set[node.Node]
}

type NoOp struct{}

func (h *Hello) ToBytes() []byte {
	nodeId := StringToBytes(h.NodeId.String())
	return AddType(HelloType, nodeId)
}

func ToHello(data []byte) *Hello {
	rawId, _ := StringFromBytes(data[1:])
	return &Hello{
		NodeId: node.IdFromRaw(rawId),
	}
}

func (h *NoOp) ToBytes() []byte {
	result := make([]byte, 1)
	result[0] = NoOpType
	return result
}

func ToNoOp(data []byte) *NoOp {
	return &NoOp{}
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

func (s *SyncNodes) GetIds() util.Set[node.Id] {
	result := util.NewSet[node.Id]()
	for _, node := range s.Nodes.GetValues() {
		result.Add(node.Id)
	}
	return result
}

func payloadType(data []byte) byte {
	if len(data) <= 0 {
		return NoOpType
	}
	return data[0]
}
