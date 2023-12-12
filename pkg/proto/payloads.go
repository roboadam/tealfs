package proto

const (
	NoOpType  = uint8(0)
	HelloType = uint8(1)
	SyncType  = uint8(2)
)

type Payload interface {
	ToBytes() []byte
}

func ToPayload(data []byte) Payload {
	switch payloadType(data) {
	case HelloType:
		return ToHello(data)
	case SyncType:
		return ToSyncNodes(data)
	default:
		return ToNoOp(data)
	}
}

func payloadType(data []byte) byte {
	if len(data) <= 0 {
		return NoOpType
	}
	return data[0]
}
