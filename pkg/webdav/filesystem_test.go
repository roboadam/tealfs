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
	"os"
	"tealfs/pkg/model"
	"tealfs/pkg/webdav"
	"testing"
)

func TestMkdir(t *testing.T) {
	inBroadcast := make(chan model.Broadcast)
	outBroadcast := make(chan model.Broadcast)
	fs := webdav.NewFileSystem(model.NewNodeId(), inBroadcast, outBroadcast)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	mockPushesAndPulls(ctx, &fs, outBroadcast)
	c := context.Background()
	mode := os.ModeDir

	err := fs.Mkdir(c, "/test", mode)
	if err != nil || !dirOpenedOk(fs, "/test") {
		t.Error("can't open dir")
		return
	}

	err = fs.Mkdir(c, "/test/stuff", mode)
	if err != nil || !dirOpenedOk(fs, "/test/stuff") {
		t.Error("can't open dir")
		return
	}

	err = fs.Mkdir(c, "/new/stuff", mode)
	if err == nil {
		t.Error("shouldn't have been able to open this one")
		return
	}
}

func TestRemoveAll(t *testing.T) {
	inBroadcast := make(chan model.Broadcast)
	outBroadcast := make(chan model.Broadcast)
	fs := webdav.NewFileSystem(model.NewNodeId(), inBroadcast, outBroadcast)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	mockPushesAndPulls(ctx, &fs, outBroadcast)
	mode := os.ModeDir

	_ = fs.Mkdir(ctx, "/test", mode)
	_ = fs.Mkdir(ctx, "/test/deleteMe", mode)
	createFileAndCheck(t, &fs, "/test/deleteMe/apple")
	_ = fs.Mkdir(ctx, "/test/deleteMe/test2", mode)
	createFileAndCheck(t, &fs, "/test/deleteMe/test2/pear")

	err := fs.RemoveAll(ctx, "/test/delete")
	if err == nil {
		t.Error("shouldn't have been able to delete this one")
		return
	}
	err = fs.RemoveAll(ctx, "/test/deleteMe")
	if err != nil {
		t.Error("should have been able to delete this one")
		return
	}
}

func TestRename(t *testing.T) {
	inBroadcast := make(chan model.Broadcast)
	outBroadcast := make(chan model.Broadcast)
	fs := webdav.NewFileSystem(model.NewNodeId(), inBroadcast, outBroadcast)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	mockPushesAndPulls(ctx, &fs, outBroadcast)
	modeDir := os.ModeDir

	createFileAndCheck(t, &fs, "/testFile")
	err := fs.Rename(ctx, "/testFile", "/testFileNew")
	if err != nil {
		t.Error("error renaming a file")
		return
	}

	err = fs.Rename(ctx, "/testFile", "/testFileNew")
	if err == nil {
		t.Error("no error renaming non existent file")
		return
	}

	_ = fs.Mkdir(ctx, "/test", modeDir)
	_ = fs.Mkdir(ctx, "/test/renameMe", modeDir)
	createFileAndCheck(t, &fs, "/test/renameMe/apple")
	_ = fs.Mkdir(ctx, "/test/renameMe/test2", modeDir)
	createFileAndCheck(t, &fs, "/test/renameMe/test2/pear")

	err = fs.Rename(ctx, "/test/renameMe", "/test/newDirName")
	if err != nil {
		t.Error("should have been able to rename the dir")
		return
	}

	dirExists(t, &fs, "/test")
	dirExists(t, &fs, "/test/newDirName")
	fileExists(t, &fs, "/test/newDirName/apple")
	dirExists(t, &fs, "/test/newDirName/test2")
	fileExists(t, &fs, "/test/newDirName/test2/pear")
}

func TestWriteAndRead(t *testing.T) {
	inBroadcast := make(chan model.Broadcast)
	outBroadcast := make(chan model.Broadcast)
	expectedData := []byte{1, 2, 3, 4, 5}
	fs := webdav.NewFileSystem(model.NewNodeId(), inBroadcast, outBroadcast)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	mockPushesAndPulls(ctx, &fs, outBroadcast)

	f, err := fs.OpenFile(context.Background(), "newFile.txt", os.O_RDWR|os.O_CREATE, 0666)
	if err != nil {
		t.Error("error creating newFile.txt")
		return
	}

	n, err := f.Write(expectedData)
	if err != nil {
		t.Error("error writing bytes")
		return
	}
	if n != 5 {
		t.Error("should have written 5 bytes")
		return
	}

	stat, err := f.Stat()
	if err != nil {
		t.Error("error stat-ing file")
		return
	}
	if stat.Size() != 5 {
		t.Error("wrong file size")
	}

	err = f.Close()
	if err != nil {
		t.Error("error closing created file")
		return
	}
	f, err = fs.OpenFile(context.Background(), "newFile.txt", os.O_RDONLY, 0666)
	if err != nil {
		t.Error("error opening file we just wrote")
		return
	}
	resultData := [5]byte{}
	n, err = f.Read(resultData[:])
	if err != nil {
		t.Error("error reading the data", err)
		return
	}
	if !bytes.Equal(expectedData, resultData[:]) {
		t.Error("got the wrong data")
		return
	}
	if n != 5 {
		t.Error("got the wrong data size")
		return
	}

	err = f.Close()
	if err != nil {
		t.Error("error closing opened file")
		return
	}
}

func fileExists(t *testing.T, fs *webdav.FileSystem, name string) {
	f := fileOrDirExists(t, fs, name)
	if f.IsDir() {
		t.Error("Error stat-ing file")
		return
	}
}

func dirExists(t *testing.T, fs *webdav.FileSystem, name string) {
	f := fileOrDirExists(t, fs, name)
	if !f.IsDir() {
		t.Error("file isn't dir")
		return
	}
}

func fileOrDirExists(t *testing.T, fs *webdav.FileSystem, name string) os.FileInfo {
	ctx := context.Background()
	info, err := fs.Stat(ctx, name)
	if err != nil {
		t.Error("error stat-ing file", err)
		return nil
	}
	return info
}

func createFileAndCheck(t *testing.T, fs *webdav.FileSystem, name string) {
	f, err := fs.OpenFile(context.Background(), name, os.O_RDWR|os.O_CREATE, 0666)
	if err != nil {
		t.Error("error creating", name)

	}
	err = f.Close()
	if err != nil {
		t.Error("error closing created file", name)
	}
	f, err = fs.OpenFile(context.Background(), name, os.O_RDONLY, 0666)
	if err != nil {
		t.Error("error opening", name)
	}
	err = f.Close()
	if err != nil {
		t.Error("error closing opened file", name)
	}
}

func dirOpenedOk(fs webdav.FileSystem, name string) bool {
	f, err := fs.OpenFile(context.Background(), name, os.O_RDONLY, os.ModeDir)
	if err != nil {
		return false
	}
	if f == nil {
		return false
	}
	return true
}
