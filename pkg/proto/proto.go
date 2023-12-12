package proto

import "tealfs/pkg/model/node"

func HelloToBytes(id node.Id) []byte {
	return StringToBytes(id.String())
}

func HelloFromBytes(bytes []byte) (node.Id, []byte) {
	value, remaining := StringFromBytes(bytes)
	return node.IdFromRaw(value), remaining
}

func NodeSyncToBytes(id node.Id) []byte {
	buffer := make([]byte, 1+len(id.String()))
	copy(buffer[1:], id.String())
	return buffer
}

func NodeSyncFromBytes(bytes []byte) (node.Id, []byte) {
	value, remaining := StringFromBytes(bytes)
	return node.IdFromRaw(value), remaining
}
