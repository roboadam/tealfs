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

package mgr

import (
	"bytes"
	"tealfs/pkg/disk"
	"tealfs/pkg/model"
	"tealfs/pkg/set"
)

type GlobalBlockListCommand struct {
	Type    GlobalBlockListCommandType
	BlockId model.BlockId
}

type GlobalBlockListCommandType uint8

type GlobalBlockList set.Set[model.BlockId]

const (
	Add    GlobalBlockListCommandType = 0
	Delete GlobalBlockListCommandType = 1
)


func (g *GlobalBlockListCommand) ToBytes() []byte {
	typeVal := model.IntToBytes(uint32(g.Type))
	blockId := model.StringToBytes(string(g.BlockId))
	return bytes.Join([][]byte{typeVal, blockId}, []byte{})
}

func (g *GlobalBlockList) SaveFile(ops disk.FileOps, path string) error {
	return ops.WriteFile(path, g.ToBytes())
}

func ToGlobalBlockListCommand(data []byte) GlobalBlockListCommand {
	typeVal, remainder := model.IntFromBytes(data)
	blockId, _ := model.StringFromBytes(remainder)
	return GlobalBlockListCommand{
		Type:    GlobalBlockListCommandType(typeVal),
		BlockId: model.BlockId(blockId),
	}
}
