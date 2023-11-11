package node

import "tealfs/pkg/proto"

func HelloToBytes(id Id) []byte {
	buffer := make([]byte, 1+len(id.String()))
	buffer[0] = proto.Hello
	copy(buffer[1:], id.String())
	return buffer
}

func HelloFromBytes(bytes []byte) (Id, []byte) {
	value, remaining := proto.StringFromBytes(bytes)
	return IdFromRaw(value), remaining
}

func NodeSyncToBytes(id Id) []byte {
	buffer := make([]byte, 1+len(id.String()))
	buffer[0] = proto.NodeSync
	copy(buffer[1:], id.String())
	return buffer
}

func NodeSyncFromBytes(bytes []byte) (Id, []byte) {
	value, remaining := proto.StringFromBytes(bytes)
	return IdFromRaw(value), remaining
}
