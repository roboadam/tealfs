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

package main

import (
	"bytes"
	"context"
	"net/http"
	"os"
	"testing"
	"time"
)

func TestOneNodeCluster(t *testing.T) {
	webdavAddress := "localhost:8080"
	uiAddress := "localhost:8081"
	nodeAddress := "localhost:8082"
	storagePath := "tmp"
	webdavUrl := "http://" + webdavAddress
	os.Mkdir(storagePath, 0755)
	defer os.RemoveAll(storagePath)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go startTealFs(storagePath, webdavAddress, uiAddress, nodeAddress, ctx)
	time.Sleep(time.Second)

	resp, ok := putFile(ctx, webdavUrl+"/test.txt", "test content", t)
	if !ok {
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		t.Error("error response", resp.Status)
		return
	}
}

func putFile(ctx context.Context, url string, contents string, t *testing.T) (*http.Response, bool) {
	client := http.Client{}
	req, err := http.NewRequestWithContext(ctx, http.MethodPut, url, bytes.NewBufferString(contents))
	if err != nil {
		t.Error("error creating request", err)
		return nil, false
	}
	req.Header.Set("Content-Type", "text/plain")

	resp, err := client.Do(req)
	if err != nil {
		t.Error("error executing request", err)
		return nil, false
	}
	return resp, true
}
