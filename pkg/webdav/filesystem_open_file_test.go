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
	"io/fs"
	"os"
	"tealfs/pkg/model"
	"tealfs/pkg/webdav"
	"testing"
)

func TestCreateEmptyFile(t *testing.T) {
	nodeId := model.NewNodeId()
	fs := webdav.NewFileSystem(nodeId)
	name := "/hello-world.txt"
	bytesInFile := []byte{1, 2, 3}
	bytesInWrite := []byte{6, 5, 4, 3, 2}
	go handleFetchBlockReq(fs.ReadReqResp, model.NewNodeId(), bytesInFile)
	go handlePushBlockReq(fs.WriteReqResp, model.NewNodeId(), bytesInWrite, t)

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
	fs := webdav.NewFileSystem(model.NewNodeId())
	go handleFetchBlockReq(fs.ReadReqResp, model.NewNodeId(), []byte{})
	_, err := fs.OpenFile(context.Background(), "/file-not-found", os.O_RDONLY, 0444)
	if err == nil {
		t.Error("Shouldn't be able to open file", err)
	}
}

func TestOpenRoot(t *testing.T) {
	filesystem := webdav.NewFileSystem(model.NewNodeId())
	root, err := filesystem.OpenFile(context.Background(), "/", os.O_RDONLY, fs.ModeDir)
	if err != nil {
		t.Error("Should be able to open root dir", err)
		return
	}
	err = root.Close()
	if err != nil {
		t.Error("Should be able to close root dir", err)
		return
	}
	rootFileInfo, err := filesystem.Stat(context.Background(), "/")
	if err != nil {
		t.Error("Should be able to stat root dir", err)
		return
	}
	if !rootFileInfo.IsDir() {
		t.Error("Root should be a dir", err)
		return
	}
}

func handleFetchBlockReq(reqs chan webdav.ReadReqResp, caller model.NodeId, bytesStored []byte) {
	for {
		req := <-reqs
		req.Resp <- model.ReadResult{
			Ok:     true,
			Caller: caller,
			Block: model.Block{
				Id:   req.Req.BlockId,
				Data: bytesStored,
			},
		}
	}
}

func handlePushBlockReq(reqs chan webdav.WriteReqResp, caller model.NodeId, expected []byte, t *testing.T) {
	for {
		req := <-reqs
		if !bytes.Equal(req.Req.Block.Data, expected) {
			t.Error("unexpected push")
		}
		req.Resp <- model.WriteResult{
			Ok:      true,
			Caller:  caller,
			BlockId: req.Req.Block.Id,
		}
	}
}
