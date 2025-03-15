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
	"errors"
	"io"
	"net/http"
	"sync"
	"tealfs/pkg/disk"
	"tealfs/pkg/model"
	"tealfs/pkg/webdav"
	"testing"
	"time"
)

func TestCreateFile(t *testing.T) {
	nodeId := model.NewNodeId()
	webdavMgrGets := make(chan model.GetBlockReq)
	webdavMgrPuts := make(chan model.PutBlockReq)
	webdavMgrBroadcast := make(chan model.Broadcast)
	mgrWebdavGets := make(chan model.GetBlockResp)
	mgrWebdavPuts := make(chan model.PutBlockResp)
	mgrWebdavBroadcast := make(chan model.Broadcast)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	mux := sync.Mutex{}
	mockStorage := make(map[model.BlockId][]byte)
	go handleWebdavMgrGets(ctx, webdavMgrGets, mgrWebdavGets, &mux, mockStorage)
	go handleWebdavMgrPuts(ctx, webdavMgrPuts, mgrWebdavPuts, &mux, mockStorage)
	go handleOutBroadcast(ctx, webdavMgrBroadcast)

	_ = webdav.New(
		nodeId,
		webdavMgrGets,
		webdavMgrPuts,
		webdavMgrBroadcast,
		mgrWebdavGets,
		mgrWebdavPuts,
		mgrWebdavBroadcast,
		"localhost:7654",
		ctx,
		&disk.MockFileOps{},
		"indexPath",
		0,
	)
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

func handleWebdavMgrGets(ctx context.Context, channel chan model.GetBlockReq, respChan chan model.GetBlockResp, mux *sync.Mutex, data map[model.BlockId][]byte) {
	for {
		select {
		case req := <-channel:
			mux.Lock()
			blockData, exists := data[req.BlockId]
			if exists {
				respChan <- model.GetBlockResp{
					Id:    req.Id(),
					Block: model.Block{Id: req.BlockId, Data: blockData},
				}
			} else {
				respChan <- model.GetBlockResp{
					Id:    req.Id(),
					Block: model.Block{Id: req.BlockId, Data: []byte{}},
				}
			}
			mux.Unlock()
		case <-ctx.Done():
			return
		}
	}
}

func handleWebdavMgrPuts(ctx context.Context, channel chan model.PutBlockReq, result chan model.PutBlockResp, mux *sync.Mutex, data map[model.BlockId][]byte) {
	for {
		select {
		case <-ctx.Done():
			return
		case req := <-channel:
			mux.Lock()
			data[req.Block.Id] = req.Block.Data
			result <- model.PutBlockResp{Id: req.Id()}
			mux.Unlock()
		}
	}
}
