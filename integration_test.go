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
	"net/url"
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

	resp, ok := putFile(ctx, webdavUrl, "text/plain", fileContents, t)
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

func TestTwoNodeCluster(t *testing.T) {
	webdavAddress1 := "localhost:8080"
	webdavAddress2 := "localhost:9080"
	uiAddress1 := "localhost:8081"
	uiAddress2 := "localhost:9081"
	nodeAddress1 := "localhost:8082"
	nodeAddress2 := "localhost:9082"
	storagePath1 := "tmp1"
	storagePath2 := "tmp2"
	connectToUrl := "http://" + uiAddress1 + "/connect-to"
	webdavUrl := "http://" + webdavAddress1 + "/test.txt"
	fileContents := "test content"
	connectToContents := "hostAndPort=" + url.QueryEscape(nodeAddress2)
	os.Mkdir(storagePath1, 0755)
	defer os.RemoveAll(storagePath1)
	os.Mkdir(storagePath2, 0755)
	defer os.RemoveAll(storagePath2)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go startTealFs(storagePath1, webdavAddress1, uiAddress1, nodeAddress1, ctx)
	go startTealFs(storagePath2, webdavAddress2, uiAddress2, nodeAddress2, ctx)
	time.Sleep(time.Second)

	resp, ok := putFile(ctx, connectToUrl, "application/x-www-form-urlencoded", connectToContents, t)
	if !ok {
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		t.Error("error response", resp.Status)
		return
	}

	resp, ok = putFile(ctx, webdavUrl, "text/plain", fileContents, t)
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

func putFile(ctx context.Context, url string, contentType string, contents string, t *testing.T) (*http.Response, bool) {
	client := http.Client{}
	req, err := http.NewRequestWithContext(ctx, http.MethodPut, url, bytes.NewBufferString(contents))
	if err != nil {
		t.Error("error creating request", err)
		return nil, false
	}
	req.Header.Set("Content-Type", contentType)

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
