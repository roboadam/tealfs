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

package webdav_test

import (
	"context"
	"os"
	"tealfs/pkg/model"
	"tealfs/pkg/webdav"
	"testing"
)

func TestMkdir(t *testing.T) {
	fs := webdav.NewFileSystem(model.NewNodeId())
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
	fs := webdav.NewFileSystem(model.NewNodeId())
	c := context.Background()
	mode := os.ModeDir

	_ = fs.Mkdir(c, "/test", mode)
	_ = fs.Mkdir(c, "/test/deleteMe", mode)
	createFileAndCheck(t, &fs, "/test/deleteMe/apple")
	_ = fs.Mkdir(c, "/test/deleteMe/test2", mode)
	createFileAndCheck(t, &fs, "/test/deleteMe/test2/pear")

	err := fs.RemoveAll(c, "/test/delete")
	if err == nil {
		t.Error("shouldn't have been able to delete this one")
		return
	}
	err = fs.RemoveAll(c, "/test/deleteMe")
	if err != nil {
		t.Error("should have been able to delete this one")
		return
	}
}

func TestRename(t *testing.T) {
	fs := webdav.NewFileSystem(model.NewNodeId())
	c := context.Background()
	mode := os.ModeDir

	createFileAndCheck(t, &fs, "/testFile")
	err := fs.Rename(c, "/testFile", "/testFileNew")
	if err != nil {
		t.Error("error renaming a file")
		return
	}

	err = fs.Rename(c, "/testFile", "/testFileNew")
	if err == nil {
		t.Error("no error renaming non existent file")
		return
	}

	_ = fs.Mkdir(c, "/test", mode)
	_ = fs.Mkdir(c, "/test/renameMe", mode)
	createFileAndCheck(t, &fs, "/test/renameMe/apple")
	_ = fs.Mkdir(c, "/test/renameMe/test2", mode)
	createFileAndCheck(t, &fs, "/test/renameMe/test2/pear")

	err = fs.Rename(c, "/test/renameMe", "/test/newDirName")
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
