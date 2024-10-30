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
	"bytes"
	"context"
	"os"
	"tealfs/pkg/webdav"
	"testing"
)

func TestCreateEmptyFile(t *testing.T) {
	fs := webdav.NewFileSystem()
	name := "/hello-world.txt"
	bytesInFile := []byte{1, 2, 3}
	bytesInWrite := []byte{6, 5, 4, 3, 2}
	go handleFetchBlockReq(fs.FetchBlockReq)
	go handlePushBlockReq(fs.PushBlockReq, bytesInFile, t)

	f, err := fs.OpenFile(context.Background(), name, os.O_RDWR|os.O_CREATE, 0666)
	if err != nil {
		t.Error("Error opening file", err)
		return
	}

	err = f.Close()
	if err != nil {
		t.Error("Error closing file", err)
		return
	}

	f, err = fs.OpenFile(context.Background(), name, os.O_RDONLY, 0444)
	if err != nil {
		t.Error("Error opening file", err)
	}

	dataRead := make([]byte, 10)
	size, err := f.Read(dataRead)
	if err != nil {
		t.Error("Error reading from file", err)
		return
	}
	if size != 3 || !bytes.Equal([]byte{1, 2, 3}, dataRead[:size]) {
		t.Error("File should be of length 3", err)
		return
	}

	numWritten, err := f.Write(bytesInWrite)
	if err != nil {
		t.Error("error pushing", err)
		return
	}
	if numWritten != len(bytesInWrite) {
		t.Error("wrong number of blocks written", err)
		return
	}
}

func TestFileNotFound(t *testing.T) {
	fs := webdav.NewFileSystem()
	go handleFetchBlockReq(fs.FetchBlockReq)
	_, err := fs.OpenFile(context.Background(), "/file-not-found", os.O_RDONLY, 0444)
	if err == nil {
		t.Error("Shouldn't be able to open file", err)
	}
}

func handleFetchBlockReq(reqs chan webdav.FetchBlockReq) {
	for {
		req := <-reqs
		req.Resp <- []byte{1, 2, 3}
	}
}

func handlePushBlockReq(reqs chan webdav.PushBlockReq, expected []byte, t *testing.T) {
	for {
		req := <-reqs
		if !bytes.Equal(req.Data, expected) {
			t.Error("unexpected push")
		}
		req.Resp <- nil
	}
}
