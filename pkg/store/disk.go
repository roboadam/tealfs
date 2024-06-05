package store

import (
	"bytes"
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

func (r *Block) Equal(o *Block) bool {
	if r.Id != o.Id {
		return false
	}
	if !bytes.Equal(r.Data, o.Data) {
		return false
	}
	if !bytes.Equal(r.Hash.Value, o.Hash.Value) {
		return false
	}

	return true
}
