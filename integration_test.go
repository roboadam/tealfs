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
	"io"
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
	webdavUrl := "http://" + webdavAddress + "/test.txt"
	fileContents := "test content"
	os.Mkdir(storagePath, 0755)
	defer os.RemoveAll(storagePath)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go startTealFs(storagePath, webdavAddress, uiAddress, nodeAddress, ctx)
	time.Sleep(time.Second)

	resp, ok := putFile(ctx, webdavUrl, fileContents, t)
	if !ok {
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		t.Error("error response", resp.Status)
		return
	}

	fetchedContent, ok := getFile(ctx, webdavUrl, t)
	if !ok {
		return
	}
	if fetchedContent != fileContents {
		t.Error("unexpected contents", resp.Status)
		return
	}
}

func getFile(ctx context.Context, url string, t *testing.T) (string, bool) {
	client := http.Client{}
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		t.Error("error creating request", err)
		return "", false
	}

	resp, err := client.Do(req)
	if err != nil {
		t.Error("error executing request", err)
		return "", false
	}
	body, err := readAllToString(resp.Body)
	if err != nil {
		t.Error("error reading body", err)
		return "", false
	}
	return body, true
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

func readAllToString(rc io.ReadCloser) (string, error) {
	defer rc.Close()
	bytes, err := io.ReadAll(rc)
	if err != nil {
		return "", err
	}
	return string(bytes), nil
}
