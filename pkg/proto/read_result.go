package proto

import (
	"bytes"
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
	block := BlockToBytes(r.Block)
	return bytes.Join([][]byte{ok, message, caller, block}, []byte{})
}

func ToReadResult(data []byte) *ReadResult {
	ok, remainder := BoolFromBytes(data)
	message, remainder := StringFromBytes(remainder)
	caller, remainder := StringFromBytes(remainder)
	block, remainder := BlockFromBytes(remainder)
	return &ReadResult{
		Ok:      ok,
		Message: message,
		Caller:  nodes.Id(caller),
		Block:   block,
	}
}
