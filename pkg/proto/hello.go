package proto

import (
	"tealfs/pkg/mgr"
	"tealfs/pkg/nodes"
)

type Hello struct {
	NodeId mgr.NodeNewId
}

func (h *Hello) ToBytes() []byte {
	nodeId := StringToBytes(h.NodeId.String())
	return AddType(HelloType, nodeId)
}

func ToHello(data []byte) *Hello {
	rawId, _ := StringFromBytes(data)
	return &Hello{
		NodeId: nodes.IdFromRaw(rawId),
	}
}
