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
	"errors"
	"fmt"
	"io"
	"net/http"
	"tealfs/pkg/model"
	"tealfs/pkg/webdav"
	"testing"
)

func TestWebdav(t *testing.T) {
	nodeId := model.NewNodeId()
	webdavMgrGets := make(chan model.ReadRequest)
	webdavMgrPuts := make(chan model.WriteRequest)
	mgrWebdavGets := make(chan model.ReadResult)
	mgrWebdavPuts := make(chan model.WriteResult)
	_ = webdav.New(nodeId, webdavMgrGets, webdavMgrPuts, mgrWebdavGets, mgrWebdavPuts, "localhost:7654")
	_, err := propFind("http://localhost:7654/")
	if err != nil {
		t.Error("error getting root", err)
		return
	}
}

func TestCreateFile(t *testing.T) {
	nodeId := model.NewNodeId()
	webdavMgrGets := make(chan model.ReadRequest)
	webdavMgrPuts := make(chan model.WriteRequest)
	mgrWebdavGets := make(chan model.ReadResult)
	mgrWebdavPuts := make(chan model.WriteResult)
	otherNode := model.NewNodeId()
	go handleWebdavMgrGets(webdavMgrGets, mgrWebdavGets, "hello world!", otherNode)
	go handleWebdavMgrPuts(t, webdavMgrPuts, mgrWebdavPuts, "hello world!", otherNode)
	_ = webdav.New(nodeId, webdavMgrGets, webdavMgrPuts, mgrWebdavGets, mgrWebdavPuts, "localhost:7654")
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
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated && resp.StatusCode != http.StatusOK {
		t.Error("status code error putting hello world", err)
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

func handleWebdavMgrGets(channel chan model.ReadRequest, respChan chan model.ReadResult, response string, caller model.NodeId) {
	for req := range channel {
		fmt.Println("get", req.BlockId, req.Caller)
		respChan <- model.ReadResult{
			Ok:     true,
			Caller: caller,
			Block: model.Block{
				Id:   req.BlockId,
				Data: []byte(response),
			},
		}
	}
}

func handleWebdavMgrPuts(t *testing.T, channel chan model.WriteRequest, result chan model.WriteResult, expectedData string, caller model.NodeId) {
	for req := range channel {
		fmt.Println("put", req.Block.Id, req.Caller)
		if string(req.Block.Data) != expectedData {
			t.Error("did not receive expected data")
			return
		}
		result <- model.WriteResult{
			Ok:      true,
			Caller:  caller,
			BlockId: req.Block.Id,
		}
	}
}
