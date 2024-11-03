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
	go handleWebdavMgrGets(webdavMgrGets)
	_ = webdav.New(nodeId, webdavMgrGets, webdavMgrPuts, mgrWebdavGets, mgrWebdavPuts, "localhost:7654")
	resp, err := http.Get("http://localhost:7654/")
	if err != nil {
		t.Error("error getting root", err)
		return
	}
	if resp.StatusCode != 200 {
		t.Error("root did not respond with 200", err)
		return
	}
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Error("error getting root body", err)
		return
	}
	fmt.Println("Body:", string(body))
}

func handleWebdavMgrGets(channel chan model.ReadRequest) {
	for req := range channel {
		fmt.Println("get", req.BlockId, req.Caller)
	}
}
