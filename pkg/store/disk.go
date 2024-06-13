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
