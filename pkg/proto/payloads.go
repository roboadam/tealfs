package proto

const (
	NoOpType     = uint8(0)
	IAmType      = uint8(1)
	SyncType     = uint8(2)
	SaveDataType = uint8(3)
)

type Payload interface {
	ToBytes() []byte
}

func ToPayload(data []byte) Payload {
	switch payloadType(data) {
	case IAmType:
		return ToHello(payloadData(data))
	case SyncType:
		return ToSyncNodes(payloadData(data))
	case SaveDataType:
		return ToSaveData(payloadData(data))
	default:
		return ToNoOp(payloadData(data))
	}
}

func payloadData(data []byte) []byte {
	return data[1:]
}

func payloadType(data []byte) byte {
	if len(data) <= 0 {
		return NoOpType
	}
	return data[0]
}
