package proto

import (
	"bytes"
	"tealfs/pkg/nodes"
	"tealfs/pkg/store"
)

type ReadRequest struct {
	Caller  nodes.Id
	BlockId store.Id
}

func (r *ReadRequest) ToBytes() []byte {
	callerId := StringToBytes(string(r.Caller))
	blockId := StringToBytes(string(r.BlockId))
	return AddType(ReadDataType, bytes.Join([][]byte{callerId, blockId}, []byte{}))
}

func ToReadData(data []byte) (*ReadRequest, []byte) {
	callerId, remainder := StringFromBytes(data)
	blockId, remainder := StringFromBytes(data)
	return &ReadRequest{
		Caller:  nodes.Id(callerId),
		BlockId: store.Id(blockId),
	}, remainder
}
