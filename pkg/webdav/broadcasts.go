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

package webdav

import (
	"bytes"
	"tealfs/pkg/model"
)

type broadcastType uint32

const (
	upsertFile = iota
	deleteFile
)

type broadcastMessage struct {
	bType   broadcastType
	file    File
	blockId model.BlockId
}

func (b *broadcastMessage) toBytes() []byte {
	bType := model.IntToBytes(uint32(b.bType))
	file := model.BytesToBytes(b.file.ToBytes())
	blockId := model.StringToBytes(string(b.blockId))
	return bytes.Join([][]byte{bType, file, blockId}, []byte{})
}

func broadcastMessageFromBytes(raw []byte, fileSystem *FileSystem) (broadcastMessage, error) {
	bType, remainder := model.IntFromBytes(raw)
	file, remainder, err := FileFromBytes(remainder, fileSystem)
	if err != nil {
		return broadcastMessage{}, err
	}
	blockId, _ := model.StringFromBytes(remainder)
	return broadcastMessage{bType: broadcastType(bType), file: file, blockId: model.BlockId(blockId)}, nil
}
