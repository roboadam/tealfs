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
	"tealfs/pkg/model"
	"testing"
	"time"
)

func TestSerializeBroadcast(t *testing.T) {
	path, _ := PathFromName("/asdf")
	msg := broadcastMessage{
		bType: deleteFile,
		file: File{
			SizeValue: 5,
			ModeValue: 4,
			Modtime:   time.Now(),
			Block: []model.Block{{
				Id: "blockId",
			}},
			HasData: []bool{true},
			Path:    path,
		},
	}
	raw := msg.toBytes()
	msg2, err := broadcastMessageFromBytes(raw, &FileSystem{})
	if err != nil {
		t.Error("error deserializing")
		return
	}
	if msg2.bType != msg.bType {
		t.Error("error with bType")
		return
	}
	if msg2.file.Block[0].Id != msg.file.Block[0].Id {
		t.Error("error with block id")
		return
	}
}
