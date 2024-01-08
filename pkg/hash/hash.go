package hash

type Hash struct {
	Value []byte
}

func FromRaw(rawHash []byte) Hash {
	return Hash{Value: rawHash}
}

func ForData(data []byte) Hash {
	return Hash{Value: data[:5]}
}
