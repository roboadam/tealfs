package node

import "tealfs/pkg/proto"

const (
	NoOpType  = uint8(0)
	HelloType = uint8(1)
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

func payloadType(data []byte) byte {
	if len(data) <= 0 {
		return NoOpType
	}
	return data[0]
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
	rawId, _ := proto.StringFromBytes(data[1:])
	return &Hello{
		NodeId: IdFromRaw(rawId),
	}
}

func (h *NoOp) ToBytes() []byte {
	return make([]byte, 0)
}

func ToNoOp(data []byte) *NoOp {
	return &NoOp{}
}
