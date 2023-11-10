package proto

import "tealfs/pkg/node"

func NodeSyncToBytes(id node.Id) []byte {
	buffer := make([]byte, 1+len(id.String()))
	buffer[0] = hello
	copy(buffer[1:], id.String())
	return buffer
}

func NodeSyncFromBytes(bytes []byte) (node.Id, []byte) {
	value, remaining := StringFromBytes(bytes)
	return node.IdFromRaw(value), remaining
}
