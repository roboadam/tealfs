package proto

import "tealfs/pkg/model/node"

type Hello struct {
	NodeId node.Id
}

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
