package proto

import (
	"tealfs/pkg/nodes"
)

type Hello struct {
	NodeId nodes.Id
}

func (h *Hello) ToBytes() []byte {
	nodeId := StringToBytes(string(h.NodeId))
	return AddType(HelloType, nodeId)
}

func ToHello(data []byte) *Hello {
	rawId, _ := StringFromBytes(data)
	return &Hello{
		NodeId: nodes.Id(rawId),
	}
}
