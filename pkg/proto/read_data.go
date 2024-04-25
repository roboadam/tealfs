package proto

import (
	"tealfs/pkg/nodes"
	"tealfs/pkg/store"
)

type ReadRequest struct {
	Caller  nodes.Id
	BlockId store.Id
}

func (r *ReadRequest) ToBytes() []byte {
	data := IntToBytes(uint32(r.BlockId))
	return AddType(ReadDataType, data)
}

func ToReadData(data []byte) (*ReadRequest, []byte) {
	id, remainder := IntFromBytes(data)
	return &ReadRequest{
		BlockId: store.Id(id),
	}, remainder
}
