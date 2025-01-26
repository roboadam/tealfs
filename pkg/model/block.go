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
	FileName string
}

func (d *DiskPointer) ToBytes() []byte {
	value := StringToBytes(string(d.NodeId))
	value = append(value, StringToBytes(d.FileName)...)
	return value
}

func ToDiskPointer(data []byte) (*DiskPointer, []byte) {
	rawId, remainder := StringFromBytes(data)
	rawFileName, remainder := StringFromBytes(remainder)
	return &DiskPointer{
		NodeId:   NodeId(rawId),
		FileName: rawFileName,
	}, remainder
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

type BlockId string

func NewBlockId() BlockId {
	idValue := uuid.New()
	return BlockId(idValue.String())
}

type BlockIdResponse struct {
	BlockId BlockId
	Err     error
}

type BlockResponse struct {
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
