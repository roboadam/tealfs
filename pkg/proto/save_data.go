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

package proto

import "tealfs/pkg/store"

type SaveData struct {
	Block store.Block
}

func (s *SaveData) ToBytes() []byte {
	// Todo: This needs to be enhanced to serialize all the Block, not just the data
	// The logic should probably live in the block and be called from here
	return AddType(SaveDataType, s.Block.Data)
}

func (s *SaveData) Equal(p Payload) bool {
	if s2, ok := p.(*SaveData); ok {
		return s2.Block.Equal(&s.Block)
	}
	return false
}

func ToSaveData(data []byte) *SaveData {
	// Todo: This needs to be enhanced to deserialize all the Block, not just the data
	// The logic should probably live in the block and be called from here
	return &SaveData{
		Block: store.Block{Data: data},
	}
}
