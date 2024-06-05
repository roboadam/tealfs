package store

import (
	"tealfs/pkg/hash"

	"github.com/google/uuid"
)

type Id string

func NewId() Id {
	idValue := uuid.New()
	return Id(idValue.String())
}

type Block struct {
	Id   Id
	Data []byte
	Hash hash.Hash
}
