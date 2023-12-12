package proto

type NoOp struct{}

func (h *NoOp) ToBytes() []byte {
	result := make([]byte, 1)
	result[0] = NoOpType
	return result
}

func ToNoOp(data []byte) *NoOp {
	return &NoOp{}
}
