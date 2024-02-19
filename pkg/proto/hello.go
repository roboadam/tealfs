package proto

import (
	"tealfs/pkg/nodes"
)

type IAm struct {
	NodeId nodes.Id
}

func (h *IAm) ToBytes() []byte {
	nodeId := StringToBytes(string(h.NodeId))
	return AddType(IAmType, nodeId)
}

func ToHello(data []byte) *IAm {
	rawId, _ := StringFromBytes(data)
	return &IAm{
		NodeId: nodes.Id(rawId),
	}
}
