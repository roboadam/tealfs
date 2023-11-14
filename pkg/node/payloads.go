package node

import "tealfs/pkg/proto"

const (
	HelloId = uint8(1)
)

type Payload interface {
	ToBytes() []byte
}

type Hello struct {
	NodeId Id
}

func (h *Hello) ToBytes() []byte {
	rawId := proto.StringToBytes(h.NodeId.value)
	return proto.PrependHeader(rawId, HelloId)
}
