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
)

type Broadcast struct {
	msg  []byte
	dest BroadcastDest
}

type BroadcastDest uint8

const (
	FileSystemDest BroadcastDest = 0
	CustodianDest  BroadcastDest = 1
)

func NewBroadcast(msg []byte, dest BroadcastDest) Broadcast {
	return Broadcast{msg: msg, dest: dest}
}

func (b *Broadcast) Msg() []byte         { return b.msg }
func (b *Broadcast) Dest() BroadcastDest { return b.dest }

func (b *Broadcast) ToBytes() []byte {
	msg := BytesToBytes(b.msg)
	dest := IntToBytes(uint32(b.dest))
	data := bytes.Join([][]byte{msg, dest}, []byte{})
	return AddType(BroadcastType, data)
}

func (b *Broadcast) Equal(p Payload) bool {
	if o, ok := p.(*Broadcast); ok {
		if !bytes.Equal(b.msg, o.msg) {
			return false
		}
		return b.dest == o.dest
	}
	return false
}

func ToBroadcast(data []byte) *Broadcast {
	msg, remainder := BytesFromBytes(data)
	dest, _ := IntFromBytes(remainder)
	return &Broadcast{msg: msg, dest: BroadcastDest(dest)}
}
