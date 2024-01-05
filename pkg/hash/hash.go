package hash

type Hash struct {
	Value []byte
}

func HashForData(data []byte) Hash {
	return Hash{Value: data[:5]}
}
