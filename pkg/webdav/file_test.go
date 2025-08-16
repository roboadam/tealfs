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

package webdav_test

import (
	"bytes"
	"context"
	"io"
	"tealfs/pkg/disk"
	"tealfs/pkg/model"
	"tealfs/pkg/webdav"
	"testing"
	"time"
)

func TestRead(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	inBroadcast := make(chan webdav.FileBroadcast, 1)
	outBroadcast := make(chan model.MgrConnsSend, 1)
	fs := webdav.NewFileSystem(model.NewNodeId(), inBroadcast, &disk.MockFileOps{}, "indexPath", 0, ctx)
	fs.OutSends = outBroadcast
	fs.Mapper = model.NewNodeConnectionMapper()

	mockPushesAndPulls(ctx, &fs, outBroadcast)

	file := webdav.File{
		SizeValue: 6,
		ModeValue: 0,
		Modtime:   time.Now(),
		Position:  0,
		HasData:   []bool{true},
		Block: []model.Block{{
			Id:   "",
			Data: []byte{1, 2, 3, 4, 5, 6},
		}},
		FileSystem: &fs,
	}

	buf := make([]byte, 4)
	n, err := file.Read(buf)
	if err != nil {
		t.Error("error reading", err)
		return
	}
	if n != 4 {
		t.Error("should have read 4 bytes instead of", n)
		return
	}
	if !bytes.Equal(buf, []byte{1, 2, 3, 4}) {
		t.Error("buffer is different", buf)
		return
	}
	n, err = file.Read(buf)
	if err != nil {
		t.Error("error reading", err)
		return
	}
	if n != 2 {
		t.Error("should have read 2 bytes instead of", n)
		return
	}
	if !bytes.Equal(buf[:2], []byte{5, 6}) {
		t.Error("buffer is different", buf)
		return
	}
	n, err = file.Read(buf)
	if err != io.EOF {
		t.Error("should have reached EOF", err, n)
		return
	}

}

func TestSeek(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	inBroadcast := make(chan webdav.FileBroadcast, 1)
	outBroadcast := make(chan model.MgrConnsSend, 1)
	fs := webdav.NewFileSystem(model.NewNodeId(), inBroadcast, &disk.MockFileOps{}, "indexPath", 0, ctx)
	fs.OutSends = outBroadcast
	fs.Mapper = model.NewNodeConnectionMapper()
	mockPushesAndPulls(ctx, &fs, outBroadcast)

	file := webdav.File{
		SizeValue: 5,
		ModeValue: 0,
		Modtime:   time.Now(),
		Position:  0,
		Block: []model.Block{{
			Id:   "",
			Data: []byte{1, 2, 3, 4, 5},
		}},
		FileSystem: &fs,
	}

	result, err := file.Seek(3, io.SeekStart)
	if err != nil {
		t.Error("error seeking", err)
	}
	if result != 3 {
		t.Error("position should be 3 instead of", result)
	}

	result, err = file.Seek(3, io.SeekStart)
	if err != nil {
		t.Error("error seeking", err)
	}
	if result != 3 {
		t.Error("second position should be 3 instead of", result)
	}

	result, err = file.Seek(3, io.SeekCurrent)
	if err != nil {
		t.Error("error seeking", err)
	}
	if result != 6 {
		t.Error("second position should be 6 instead of", result)
	}

	result, err = file.Seek(-4, io.SeekEnd)
	if err != nil {
		t.Error("error seeking", err)
	}
	if result != 1 {
		t.Error("second position should be 1 instead of", result)
	}
	_, err = file.Seek(-4, io.SeekCurrent)
	if err == nil {
		t.Error("position shouldn't be allowed to be negative")
	}
}

func TestSerialize(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	nodeId := model.NewNodeId()
	fileSystem := webdav.NewFileSystem(
		nodeId,
		make(chan webdav.FileBroadcast),
		&disk.MockFileOps{},
		"indexPath",
		0,
		ctx,
	)
	fileSystem.OutSends = make(chan model.MgrConnsSend)
	fileSystem.Mapper = model.NewNodeConnectionMapper()

	path, _ := webdav.PathFromName("/hello/world")
	file := webdav.File{
		SizeValue: 123,
		ModeValue: 234,
		Modtime:   time.Unix(123456, 0),
		Position:  345,
		Block: []model.Block{{
			Id:   model.NewBlockId(),
			Data: []byte{4, 5, 6},
		}},
		HasData:    []bool{true},
		Path:       path,
		FileSystem: &fileSystem,
	}

	fileBytes := file.ToBytes()
	fileClone, _, err := webdav.FileFromBytes(fileBytes, &fileSystem)
	if err != nil {
		t.Error("error serializing file", err)
		return
	}

	if file.Block[0].Id != fileClone.Block[0].Id {
		t.Error("block id is different", file.Block[0].Id, fileClone.Block[0].Id)
	}
	if file.Name() != fileClone.Name() {
		t.Error("name is different", file.Name(), fileClone.Name())
	}
	if file.Size() != fileClone.Size() {
		t.Error("size is different", file.Size(), fileClone.Size())
	}
	if file.Mode() != fileClone.Mode() {
		t.Error("mode is different", file.Mode(), fileClone.Mode())
	}
	if file.ModTime() != fileClone.ModTime() {
		t.Error("modtime is different", file.ModTime(), fileClone.ModTime())
	}
	cancel()
}
