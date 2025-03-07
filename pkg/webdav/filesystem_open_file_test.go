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
	"context"
	"io"
	"io/fs"
	"os"
	"sync"
	"tealfs/pkg/model"
	"tealfs/pkg/webdav"
	"testing"
)

func TestCreateEmptyFile(t *testing.T) {
	nodeId := model.NewNodeId()
	fs := webdav.NewFileSystem(nodeId)
	name := "/hello-world.txt"
	bytesInWrite := []byte{6, 5, 4, 3, 2}
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	mockPushesAndPulls(ctx, &fs)

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
	_, err = f.Read(dataRead)
	if err == nil || err != io.EOF {
		t.Error("expected EOF", err)
		return
	}
	numWritten, err := f.Write(bytesInWrite)
	if err != nil {
		t.Error("error pushing", err)
		return
	}
	if numWritten != len(bytesInWrite) {
		t.Error("wrong number of blocks written. expected", len(bytesInWrite), "got", numWritten)
		return
	}
}

func TestFileNotFound(t *testing.T) {
	fs := webdav.NewFileSystem(model.NewNodeId())
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	mockPushesAndPulls(ctx, &fs)
	_, err := fs.OpenFile(context.Background(), "/file-not-found", os.O_RDONLY, 0444)
	if err == nil {
		t.Error("Shouldn't be able to open file", err)
	}
}

func TestOpenRoot(t *testing.T) {
	filesystem := webdav.NewFileSystem(model.NewNodeId())
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	mockPushesAndPulls(ctx, &filesystem)
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

func handleFetchBlockReq(ctx context.Context, reqs chan webdav.ReadReqResp, mux *sync.Mutex, data map[model.BlockId][]byte) {
	for {
		select {
		case <-ctx.Done():
			return
		case req := <-reqs:
			mux.Lock()
			blockData, exists := data[req.Req.BlockId]
			if exists {
				req.Resp <- model.GetBlockResp{
					Id: req.Req.Id(),
					Block: model.Block{
						Id:   req.Req.BlockId,
						Type: model.Mirrored,
						Data: blockData,
					},
					Err: nil,
				}
			} else {
				req.Resp <- model.GetBlockResp{
					Id: req.Req.Id(),
					Block: model.Block{
						Id:   req.Req.BlockId,
						Type: model.Mirrored,
						Data: []byte{},
					},
					Err: nil,
				}
			}
			mux.Unlock()
		}
	}
}

func handlePushBlockReq(ctx context.Context, reqs chan webdav.WriteReqResp, mux *sync.Mutex, data map[model.BlockId][]byte) {
	for {
		select {
		case <-ctx.Done():
			return
		case req := <-reqs:
			mux.Lock()
			data[req.Req.Block.Id] = req.Req.Block.Data
			req.Resp <- model.PutBlockResp{Id: req.Req.Id()}
			mux.Unlock()
		}
	}
}

func mockPushesAndPulls(ctx context.Context, fs *webdav.FileSystem) {
	mux := sync.Mutex{}
	mockStorage := make(map[model.BlockId][]byte)
	go handleFetchBlockReq(ctx, fs.ReadReqResp, &mux, mockStorage)
	go handlePushBlockReq(ctx, fs.WriteReqResp, &mux, mockStorage)
}
