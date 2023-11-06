package proto

import "tealfs/pkg/node"

func HelloToBytes(id node.Id) []byte {
	buffer := make([]byte, 1+len(id.String()))
	buffer[0] = hello
	copy(buffer[1:], id.String())
	return buffer
}

func HelloFromBytes(data []byte) node.Id {
	return node.IdFromRaw(string(data))
}
