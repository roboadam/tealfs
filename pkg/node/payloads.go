package node

import "tealfs/pkg/proto"

const (
	HelloType = uint8(1)
)

type Payload interface {
	ToBytes() []byte
}

func ToPayload(data []byte) Payload {
	switch data[0] {
	case HelloType:
		return ToHello(data)
	default:
		return ToNoOp(data)
	}
}

type Hello struct {
	NodeId Id
}

type NoOp struct{}

func (h *Hello) ToBytes() []byte {
	nodeId := proto.StringToBytes(h.NodeId.value)
	return proto.AddType(HelloType, nodeId)
}

func ToHello(data []byte) *Hello {
	return &Hello{
		NodeId: IdFromRaw(string(data[1:])),
	}
}

func (h *NoOp) ToBytes() []byte {
	return make([]byte, 0)
}

func ToNoOp(data []byte) *NoOp {
	return &NoOp{}
}
