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
	"tealfs/pkg/webdav"
	"testing"
)

func TestMkdir(t *testing.T) {
	fs := webdav.NewFileSystem()
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
	fs := webdav.NewFileSystem()
	c := context.Background()
	mode := os.ModeDir

	_ = fs.Mkdir(c, "/test", mode)
	_ = fs.Mkdir(c, "/test/deleteme", mode)
	createFileAndCheck(t, &fs, "/test/deleteme/asdf")
	_ = fs.Mkdir(c, "/test/deleteme/test2", mode)
	createFileAndCheck(t, &fs, "/test/deleteme/test2/qwer")

	err := fs.RemoveAll(c, "/test/delete")
	if err == nil {
		t.Error("shouldn't have been able to delete this one")
		return
	}
	err = fs.RemoveAll(c, "/test/deleteme")
	if err != nil {
		t.Error("should have been able to delete this one")
		return
	}
}

func TestRename(t *testing.T) {
	fs := webdav.NewFileSystem()
	c := context.Background()
	mode := os.ModeDir

	createFileAndCheck(t, &fs, "/testfile")
	err := fs.Rename(c, "/testfile", "/testfilenew")
	if err != nil {
		t.Error("error renaming a file")
		return
	}

	err = fs.Rename(c, "/testfile", "/testfilenew")
	if err == nil {
		t.Error("no error renaming non existent file")
		return
	}

	_ = fs.Mkdir(c, "/test", mode)
	_ = fs.Mkdir(c, "/test/renameme", mode)
	createFileAndCheck(t, &fs, "/test/renameme/asdf")
	_ = fs.Mkdir(c, "/test/renameme/test2", mode)
	createFileAndCheck(t, &fs, "/test/renameme/test2/qwer")

	err = fs.Rename(c, "/test/renameme", "/test/newdirname")
	if err != nil {
		t.Error("should have been able to rename the dir")
		return
	}

	dirExists(t, &fs, "/test")
	dirExists(t, &fs, "/test/newdirname")
	fileExists(t, &fs, "/test/newdirname/asdf")
	dirExists(t, &fs, "/test/newdirname/test2")
	fileExists(t, &fs, "/test/newdirname/test2/qwer")
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
		t.Error("file is dir")
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
