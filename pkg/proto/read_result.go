package proto

import (
	"tealfs/pkg/nodes"
	"tealfs/pkg/store"
)

type ReadResult struct {
	Ok      bool
	Message string
	Caller  nodes.Id
	Block   store.Block
}

func (r *ReadResult) ToBytes() []byte {
	ok := BoolToBytes(r.Ok)
	message := StringToBytes(r.Message)
	caller := StringToBytes(string(r.Caller))
	block := r.Block
}

func ToReadResult(data []byte) *ReadResult {

}
