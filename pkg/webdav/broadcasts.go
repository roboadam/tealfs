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
	bType broadcastType
	file  File
}

func (b *broadcastMessage) toBytes() []byte {
	bType := model.IntToBytes(uint32(b.bType))
	file := model.BytesToBytes(b.file.ToBytes())
	return bytes.Join([][]byte{bType, file}, []byte{})
}

func broadcastMessageFromBytes(raw []byte, fileSystem *FileSystem) (broadcastMessage, error) {
	bType, remainder := model.IntFromBytes(raw)
	rawFile, _ := model.BytesFromBytes(remainder)
	file, _, err := FileFromBytes(rawFile, fileSystem)
	if err != nil {
		return broadcastMessage{}, err
	}
	return broadcastMessage{bType: broadcastType(bType), file: file}, nil
}
