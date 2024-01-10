package hash

import (
	"crypto/sha256"
)

type Hash struct {
	Value []byte
}

func FromRaw(rawHash []byte) Hash {
	return Hash{Value: rawHash}
}

func ForData(data []byte) Hash {
	value := sha256.Sum256(data)
	return Hash{value[:]}
}
