package node

import (
	"tealfs/pkg/proto"
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
	NodeId Id
}

type SyncNodes struct {
	Nodes util.Set[Node]
}

type NoOp struct{}

func (h *Hello) ToBytes() []byte {
	nodeId := proto.StringToBytes(h.NodeId.value)
	return proto.AddType(HelloType, nodeId)
}

func ToHello(data []byte) *Hello {
	rawId, _ := proto.StringFromBytes(data[1:])
	return &Hello{
		NodeId: IdFromRaw(rawId),
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
		id := proto.StringToBytes(node.Id.String())
		address := proto.StringToBytes(node.Address.Value)
		result = append(result, id...)
		result = append(result, address...)
	}
	return proto.AddType(SyncType, result)
}

func ToSyncNodes(data []byte) *SyncNodes {
	remainder := data
	result := util.NewSet[Node]()
	for {
		var id, address string
		id, remainder = proto.StringFromBytes(remainder)
		address, remainder = proto.StringFromBytes(remainder)
		node := Node{Id: IdFromRaw(id), Address: NewAddress(address)}
		result.Add(node)
		if len(remainder) <= 0 {
			return &SyncNodes{Nodes: result}
		}
	}
}

func payloadType(data []byte) byte {
	if len(data) <= 0 {
		return NoOpType
	}
	return data[0]
}
