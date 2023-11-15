package node

import "tealfs/pkg/proto"

const (
	HelloType = uint8(1)
)

type Payload interface {
	ToBytes() []byte
}

func ToPayload(data []byte) *Payload {
	switch data[0] {
		case HelloType
	}
}

type Hello struct {
	NodeId Id
}

func (h *Hello) ToBytes() []byte {
	nodeId := proto.StringToBytes(h.NodeId.value)
	return proto.AddType(HelloType, nodeId)
}

