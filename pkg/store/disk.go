package store

import "tealfs/pkg/hash"

type Id int

type Block struct {
	Id       Id
	Data     []byte
	Hash     hash.Hash
	Children []Id
}
