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
	"errors"
	"io"
	"net/http"
	"sync"
	"tealfs/pkg/model"
	"tealfs/pkg/webdav"
	"testing"
	"time"
)

func TestCreateFile(t *testing.T) {
	nodeId := model.NewNodeId()
	webdavMgrGets := make(chan model.ReadRequest)
	webdavMgrPuts := make(chan model.WriteRequest)
	mgrWebdavGets := make(chan model.ReadResult)
	mgrWebdavPuts := make(chan model.WriteResult)
	mgrWebdavLockMsg := make(chan webdav.LockMessage)
	webdavMgrLockMsg := make(chan webdav.LockMessage)
	otherNode := model.NewNodeId()
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	mux := sync.Mutex{}
	mockStorage := make(map[model.BlockId][]byte)
	go handleWebdavMgrGets(ctx, webdavMgrGets, mgrWebdavGets, otherNode, &mux, mockStorage)
	go handleWebdavMgrPuts(ctx, webdavMgrPuts, mgrWebdavPuts, otherNode, &mux, mockStorage)

	_ = webdav.New(nodeId, webdavMgrGets, webdavMgrPuts, mgrWebdavGets, mgrWebdavPuts, webdavMgrLockMsg, mgrWebdavLockMsg, "localhost:7654", ctx)
	time.Sleep(1 * time.Second) //FIXME, need a better way to wait for listener to start

	_, err := propFind("http://localhost:7654/")
	if err != nil {
		t.Error("error getting root", err)
		cancel()
		return
	}

	url := "http://localhost:7654/hello_world.txt"
	content := []byte("hello world!")

	req, _ := http.NewRequest("PUT", url, bytes.NewBuffer(content))
	req.Header.Set("Content-Type", "text/plain")
	client := &http.Client{}
	resp, err := client.Do(req)

	if err != nil {
		t.Error("error putting hello world", err)
		return
	}
	resp.Body.Close()

	if resp.StatusCode != http.StatusCreated && resp.StatusCode != http.StatusOK {
		t.Error("status code error putting hello world", err)
		return
	}

	resp, err = http.Get(url)
	if err != nil {
		t.Error("error getting hello world", err)
		return
	}
	if resp.StatusCode != http.StatusOK {
		t.Error("status code error getting hello world", err)
		return
	}
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Error("error reading body", err)
		return
	}
	if string(body) != "hello world!" {
		t.Error("body not expected", string(body))
		return
	}
}

func propFind(url string) (string, error) {
	req, err := http.NewRequest("PROPFIND", url, nil)
	if err != nil {
		return "", err
	}
	req.Header.Set("Depth", "1")
	req.Header.Set("Content-Type", "application/xml")
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	if resp.StatusCode < 200 || resp.StatusCode >= 400 {
		return "", errors.New("status not successful " + resp.Status)
	}

	return string(body), nil
}

func handleWebdavMgrGets(ctx context.Context, channel chan model.ReadRequest, respChan chan model.ReadResult, caller model.NodeId, mux *sync.Mutex, data map[model.BlockId][]byte) {
	for {
		select {
		case req := <-channel:
			mux.Lock()
			blockData, exists := data[req.BlockId]
			if exists {
				respChan <- model.ReadResult{
					Ok:     true,
					Caller: caller,
					Block: model.Block{
						Id:   req.BlockId,
						Data: blockData,
					},
				}
			} else {
				respChan <- model.ReadResult{
					Ok:     true,
					Caller: caller,
					Block: model.Block{
						Id:   req.BlockId,
						Data: []byte{},
					},
				}
			}
			mux.Unlock()
		case <-ctx.Done():
			return
		}
	}
}

func handleWebdavMgrPuts(ctx context.Context, channel chan model.WriteRequest, result chan model.WriteResult, caller model.NodeId, mux *sync.Mutex, data map[model.BlockId][]byte) {
	for {
		select {
		case <-ctx.Done():
			return
		case req := <-channel:
			mux.Lock()
			data[req.Block.Id] = req.Block.Data
			result <- model.WriteResult{
				Ok:      true,
				Caller:  caller,
				BlockId: req.Block.Id,
			}
			mux.Unlock()
		}
	}
}
