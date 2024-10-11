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
	go handleFetchBlockReq(fs.FetchBlockReq)

	f, err := fs.OpenFile(context.Background(), name, os.O_RDWR|os.O_CREATE, 0666)
	if err != nil {
		t.Error("Error opening file", err)
		return
	}

	err = f.Close()
	if err != nil {
		t.Error("Error closing file", err)
	}

	f, err = fs.OpenFile(context.Background(), name, os.O_RDONLY, 0444)
	if err != nil {
		t.Error("Error opening file", err)
	}

	dataRead := make([]byte, 10)
	len, err := f.Read(dataRead)
	if err != nil {
		t.Error("Error reading from file", err)
	}
	if len != 3 || !bytes.Equal([]byte{1, 2, 3}, dataRead[:len]) {
		t.Error("File shoud be of length 3", err)
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
