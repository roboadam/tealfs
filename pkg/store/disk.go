package store

import "tealfs/pkg/hash"

type Id string

type Block struct {
	Id       Id
	Data     []byte
	Hash     hash.Hash
	Children []Id
}
