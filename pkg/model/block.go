// Copyright (C) 2025 Adam Hess
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

	"github.com/google/uuid"
)

type BlockType uint32

const (
	Mirrored BlockType = iota
	XORed
)

type RawData struct {
	Ptr  DiskPointer
	Data []byte
}

func ToRawData(dataRaw []byte) (*RawData, []byte) {
	ptr, remainder := ToDiskPointer(dataRaw)
	data, remainder := BytesFromBytes(remainder)
	return &RawData{
		Ptr:  *ptr,
		Data: data,
	}, remainder
}

func (b *RawData) ToBytes() []byte {
	ptr := b.Ptr.ToBytes()
	data := BytesToBytes(b.Data)
	return bytes.Join([][]byte{ptr, data}, []byte{})
}

func (b *RawData) Equals(o *RawData) bool {
	if !b.Ptr.Equals(&o.Ptr) {
		return false
	}
	if !bytes.Equal(b.Data, o.Data) {
		return false
	}
	return true
}

type DiskPointer struct {
	NodeId   NodeId
	Disk     DiskId
	FileName string
}

type GetBlockReq struct {
	id      GetBlockId
	BlockId BlockId
}

type PutBlockReq struct {
	id    PutBlockId
	Block Block
}

func NewGetBlockReq(blockId BlockId) GetBlockReq {
	id := GetBlockId(uuid.New().String())
	return GetBlockReq{id, blockId}
}

func NewPutBlockReq(block Block) PutBlockReq {
	id := PutBlockId(uuid.New().String())
	return PutBlockReq{id, block}
}

func (g *GetBlockReq) Id() GetBlockId {
	return g.id
}

func (p *PutBlockReq) Id() PutBlockId {
	return p.id
}

type GetBlockId string
type PutBlockId string

type BlockId string

func NewBlockId() BlockId {
	idValue := uuid.New()
	return BlockId(idValue.String())
}

type PutBlockResp struct {
	Id  PutBlockId
	Err error
}

type GetBlockResp struct {
	Id    GetBlockId
	Block Block
	Err   error
}

type Block struct {
	Id   BlockId
	Type BlockType
	Data []byte
}

func (r *Block) Equal(o *Block) bool {
	if r.Id != o.Id {
		return false
	}
	if r.Type != o.Type {
		return false
	}
	if !bytes.Equal(r.Data, o.Data) {
		return false
	}

	return true
}
