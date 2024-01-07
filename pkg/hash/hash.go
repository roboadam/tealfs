package hash

type Hash struct {
	Value []byte
}

func HashFromRaw(rawHash []byte) Hash {
	return Hash{Value: rawHash}
}

func HashForData(data []byte) Hash {
	return Hash{Value: data[:5]}
}
