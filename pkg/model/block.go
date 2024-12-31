// Copyright (C) 2024 Adam Hess
//
// This program is free software: you can redistribute it and/or modify it under
// the terms of the GNU Affero General Public License as published by the Free
// Software Foundation, version 3.
//
// This program is distributed in the hope that it will be useful, but WITHOUT
// ANY WARRANTY; without even the implied warranty of MERCHANTABILITY or
// FITNESS FOR A PARTICULAR PURPOSE. See the GNU Affero General Public License
// for more details.
//
// You should have received a copy of the GNU Affero General Public License
// along with this program. If not, see <https://www.gnu.org/licenses/>.

package model

import (
	"bytes"
)

type BlockKeyType uint32
type BlockKeyId string

const (
	Mirrored BlockKeyType = iota
	XORed
)

type BlockKey struct {
	Id     BlockKeyId
	Type   BlockKeyType
	Data   []DiskPointer
	Parity DiskPointer
}

func (b *BlockKey) ToBytes() []byte {
	value := StringToBytes(string(b.Id))
	value = append(value, IntToBytes(uint32(b.Type))...)
	for i := range b.Data {
		value = append(value, BytesToBytes(b.Data[i].ToBytes())...)
	}
	value = append(value, b.Parity.ToBytes()...)
	return value
}

func (b *BlockKey) Equals(o *BlockKey) bool {
	if b.Id != o.Id {
		return false
	}
	if b.Type != o.Type {
		return false
	}
	if len(b.Data) != len(o.Data) {
		return false
	}
	for i := range b.Data {
		if !b.Data[i].Equals(&o.Data[i]) {
			return false
		}
	}
	if !b.Parity.Equals(&o.Parity) {
		return false
	}
	return true
}

type DiskPointer struct {
	NodeId   NodeId
	FileName string
}

func (d *DiskPointer) ToBytes() []byte {
	value := StringToBytes(string(d.NodeId))
	value = append(value, StringToBytes(d.FileName)...)
	return value
}

func (d *DiskPointer) Equals(o *DiskPointer) bool {
	if d.NodeId != o.NodeId {
		return false
	}
	if d.FileName != o.FileName {
		return false
	}
	return true
}

// type BlockId string

// func NewBlockId() BlockId {
// 	idValue := uuid.New()
// 	return BlockId(idValue.String())
// }

type Block struct {
	Id   BlockKey
	Data []byte
}

func (r *Block) Equal(o *Block) bool {
	if !r.Id.Equals(&o.Id) {
		return false
	}
	if !bytes.Equal(r.Data, o.Data) {
		return false
	}

	return true
}
