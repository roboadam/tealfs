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

package main

import (
	"bytes"
	"context"
	"io"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/sirupsen/logrus"
)

func TestOneNodeCluster(t *testing.T) {
	webdavAddress := "localhost:7080"
	uiAddress := "localhost:7081"
	nodeAddress := "localhost:7082"
	storagePath := "tmp"
	webdavUrl := "http://" + webdavAddress + "/test.txt"
	fileContents := "You will rejoice to hear that no disaster has accompanied the commencement of an enterprise which you have regarded with such evil forebodings. I arrived here yesterday, and my first task is to assure my dear sister of my welfare and increasing confidence in the success of my undertaking. I am already far north of London, and as I walk in the streets of Petersburgh, I feel a cold northern breeze play upon my cheeks, which braces my nerves and fills me with delight. Do you understand this feeling? This breeze, which has travelled from the regions towards which I am advancing, gives me a foretaste of those icy climes. Inspirited by this wind of promise, my daydreams become more fervent and vivid. I try in vain to be persuaded that the pole is the seat of frost and desolation; it ever presents itself to my imagination as the region of beauty and delight. There, Margaret, the sun is for ever visible, its broad disk just skirting the horizon and diffusing a perpetual splendour. There—for with your leave, my sister, I will put some trust in preceding navigators—there snow and frost are banished; and, sailing over a calm sea, we may be wafted to a land surpassing in wonders and in beauty every region hitherto discovered on the habitable globe. Its productions and features may be without example, as the phenomena of the heavenly bodies undoubtedly are in those undiscovered solitudes. What may not be expected in a country of eternal light? I may there discover the wondrous power which attracts the needle and may regulate a thousand celestial observations that require only this voyage to render their seeming eccentricities consistent for ever. I shall satiate my ardent curiosity with the sight of a part of the world never before visited, and may tread a land never before imprinted by the foot of man. These are my enticements, and they are sufficient to conquer all fear of danger or death and to induce me to commence this laborious voyage with the joy a child feels when he embarks in a little boat, with his holiday mates, on an expedition of discovery up his native river. But supposing all these conjectures to be false, you cannot contest the inestimable benefit which I shall confer on all mankind, to the last generation, by discovering a passage near the pole to those countries, to reach which at present so many months are requisite; or by ascertaining the secret of the magnet, which, if at all possible, can only be effected by an undertaking such as mine."
	os.RemoveAll(storagePath)
	os.Mkdir(storagePath, 0755)
	defer os.RemoveAll(storagePath)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go startTealFs(storagePath, webdavAddress, uiAddress, nodeAddress, 1, ctx)
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
		t.Error("unexpected contents got:", fetchedContent, "expected:", fileContents)
		return
	}
	cancel()
}

func TestTwoNodeCluster(t *testing.T) {
	webdavAddress1 := "localhost:8080"
	webdavAddress2 := "localhost:9080"
	path1 := "/test1.txt"
	path2 := "/test2.txt"
	uiAddress1 := "localhost:8081"
	uiAddress2 := "localhost:9081"
	nodeAddress1 := "localhost:8082"
	nodeAddress2 := "localhost:9082"
	storagePath1 := "tmp1"
	storagePath2 := "tmp2"
	os.RemoveAll(storagePath1)
	os.RemoveAll(storagePath2)
	connectToUrl := "http://" + uiAddress1 + "/connect-to"
	fileContents1 := "test content 1"
	fileContents2 := "test content 2"
	connectToContents := "hostAndPort=" + url.QueryEscape(nodeAddress2)
	os.Mkdir(storagePath1, 0755)
	defer os.RemoveAll(storagePath1)
	os.Mkdir(storagePath2, 0755)
	defer os.RemoveAll(storagePath2)
	ctx1, cancel1 := context.WithCancel(context.Background())
	ctx2, cancel2 := context.WithCancel(context.Background())
	defer cancel1()
	defer cancel2()

	go startTealFs(storagePath1, webdavAddress1, uiAddress1, nodeAddress1, 1, ctx1)
	go startTealFs(storagePath2, webdavAddress2, uiAddress2, nodeAddress2, 1, ctx2)

	time.Sleep(time.Second)

	resp, ok := putFile(ctx1, connectToUrl, "application/x-www-form-urlencoded", connectToContents, t)
	if !ok {
		t.Error("error response", resp.Status)
		return
	}
	resp.Body.Close()
	time.Sleep(time.Second)

	if resp.StatusCode >= 400 {
		t.Error("error response", resp.Status)
		return
	}

	resp, ok = putFile(ctx1, urlFor(webdavAddress1, path1), "text/plain", fileContents1, t)
	if !ok {
		t.Error("error response", resp.Status)
		return
	}
	resp.Body.Close()

	if resp.StatusCode >= 400 {
		t.Error("error response", resp.Status)
		return
	}

	resp, ok = putFile(ctx2, urlFor(webdavAddress2, path2), "text/plain", fileContents2, t)
	if !ok {
		t.Error("error putting file")
		return
	}
	resp.Body.Close()

	if resp.StatusCode >= 400 {
		t.Error("error response", resp.Status)
		return
	}

	fetchedContent, ok := getFile(ctx2, urlFor(webdavAddress2, path1), t)
	if !ok {
		t.Error("error getting file")
		return
	}
	if fetchedContent != fileContents1 {
		t.Error("unexpected contents", fetchedContent)
		return
	}

	fetchedContent, ok = getFile(ctx1, urlFor(webdavAddress1, path2), t)
	if !ok {
		t.Error("error getting file")
		return
	}
	if fetchedContent != fileContents2 {
		t.Error("unexpected contents", fetchedContent)
		return
	}

	cancel1()
	time.Sleep(time.Second)
	ctx1, cancel1 = context.WithCancel(context.Background())
	defer cancel1()

	go startTealFs(storagePath1, webdavAddress1, uiAddress1, nodeAddress1, 1, ctx1)

	time.Sleep(time.Second)

	if resp.StatusCode >= 400 {
		t.Error("error response", resp.Status)
		return
	}

	fetchedContent, ok = getFile(ctx1, urlFor(webdavAddress1, path1), t)
	if !ok {
		t.Error("error getting file")
		return
	}
	if fetchedContent != fileContents1 {
		t.Error("unexpected contents", fetchedContent)
		return
	}

	fetchedContent, ok = getFile(ctx2, urlFor(webdavAddress2, path2), t)
	if !ok {
		t.Error("error getting file")
		return
	}
	if fetchedContent != fileContents2 {
		t.Error("unexpected contents", fetchedContent)
		return
	}

	uiContents1, ok := getFile(ctx1, urlFor(uiAddress1, "/"), t)
	if !ok {
		t.Error("error getting ui contents")
		return
	}

	if strings.Count(uiContents1, nodeAddress2) != 1 {
		t.Error("should be connected to remote node exactly once")
		return
	}

	if strings.Count(uiContents1, nodeAddress1) != 0 {
		t.Error("should not be connected to yourself")
		return
	}

	uiContents2, ok := getFile(ctx2, urlFor(uiAddress2, "/"), t)
	if !ok {
		t.Error("error getting ui contents")
		return
	}

	if strings.Count(uiContents2, nodeAddress1) != 1 {
		t.Error("should be connected to remote node exactly once, got:", strings.Count(uiContents2, nodeAddress1))
		return
	}

	if strings.Count(uiContents2, nodeAddress2) != 0 {
		t.Error("should not be connected to yourself")
		return
	}
	cancel1()
	cancel2()
}

func TestTwoNodeClusterLotsOfFiles(t *testing.T) {
	webdavAddress1 := "localhost:8080"
	parallel := 100
	paths := make([]string, parallel)
	fileContents := make([]string, parallel)
	for i := range parallel {
		paths[i] = "/test" + strconv.Itoa(i) + ".txt"
		fileContents[i] = "test content " + strconv.Itoa(i)
	}
	uiAddress1 := "localhost:8081"
	nodeAddress1 := "localhost:8082"
	storagePath1 := "tmp1"
	os.RemoveAll(storagePath1)
	os.Mkdir(storagePath1, 0755)
	defer os.RemoveAll(storagePath1)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go startTealFs(storagePath1, webdavAddress1, uiAddress1, nodeAddress1, 1, ctx)

	time.Sleep(time.Second)

	var wg sync.WaitGroup
	for i := range parallel {
		wg.Add(1)
		go putFileWg(paths[i], fileContents[i], &wg, t, ctx, webdavAddress1)
	}
	wg.Wait()

	wg = sync.WaitGroup{}
	for i := range parallel {
		wg.Add(1)
		go getFileWg(paths[i], fileContents[i], &wg, t, ctx, webdavAddress1)
	}
	wg.Wait()
	cancel()
}

func putFileWg(path string, contents string, wg *sync.WaitGroup, t *testing.T, ctx context.Context, webdavAddress string) {
	defer wg.Done()
	resp, ok := putFile(ctx, urlFor(webdavAddress, path), "text/plain", contents, t)
	if !ok {
		t.Error("error response", resp.Status)
		return
	}
	resp.Body.Close()

	if resp.StatusCode >= 400 {
		t.Error("error response", resp.Status)
		return
	}
}

func getFileWg(path string, expectedContents string, wg *sync.WaitGroup, t *testing.T, ctx context.Context, webdavAddress string) {
	defer wg.Done()
	fetchedContent, ok := getFile(ctx, urlFor(webdavAddress, path), t)
	if !ok {
		t.Error("error getting file", path)
		return
	}
	if fetchedContent != expectedContents {
		t.Error("for ", path, " unexpected contents:", fetchedContent, ":")
		return
	}
}

func TestTwoNodeOneStorageCluster(t *testing.T) {
	webdavAddress1 := "localhost:8080"
	webdavAddress2 := "localhost:9080"
	path1 := "/test1.txt"
	path2 := "/test2.txt"
	uiAddress1 := "localhost:8081"
	uiAddress2 := "localhost:9081"
	nodeAddress1 := "localhost:8082"
	nodeAddress2 := "localhost:9082"
	storagePath1 := "xtmp1"
	storagePath2 := "xtmp2"
	os.RemoveAll(storagePath1)
	os.RemoveAll(storagePath2)
	connectToUrl := "http://" + uiAddress1 + "/connect-to"
	fileContents1 := "test content 1"
	fileContents2 := "test content 2"
	connectToContents := "hostAndPort=" + url.QueryEscape(nodeAddress2)
	os.Mkdir(storagePath1, 0755)
	defer os.RemoveAll(storagePath1)
	os.Mkdir(storagePath2, 0755)
	defer os.RemoveAll(storagePath2)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go startTealFs(storagePath1, webdavAddress1, uiAddress1, nodeAddress1, 1, ctx)
	go startTealFs(storagePath2, webdavAddress2, uiAddress2, nodeAddress2, 1, ctx)
	time.Sleep(time.Second)

	logrus.Info("1")
	resp, ok := putFile(ctx, connectToUrl, "application/x-www-form-urlencoded", connectToContents, t)
	logrus.Info("2")
	if !ok {
		t.Error("error response", resp.Status)
		return
	}
	resp.Body.Close()
	time.Sleep(time.Second)

	if resp.StatusCode >= 400 {
		t.Error("error response", resp.Status)
		return
	}

	logrus.Info("3a")
	resp, ok = putFile(ctx, urlFor(webdavAddress1, path1), "text/plain", fileContents1, t)
	logrus.Info("3")
	if !ok {
		t.Error("error response", resp.Status)
		return
	}
	resp.Body.Close()

	if resp.StatusCode >= 400 {
		t.Error("error response", resp.Status)
		return
	}

	resp, ok = putFile(ctx, urlFor(webdavAddress2, path2), "text/plain", fileContents2, t)
	logrus.Info("4")
	if !ok {
		t.Error("error putting file")
		return
	}
	resp.Body.Close()

	if resp.StatusCode >= 400 {
		t.Error("error response", resp.Status)
		return
	}

	fetchedContent, ok := getFile(ctx, urlFor(webdavAddress2, path1), t)
	logrus.Info("5")
	if !ok {
		t.Error("error getting file")
		return
	}
	if fetchedContent != fileContents1 {
		t.Error("unexpected contents", fetchedContent)
		return
	}

	fetchedContent, ok = getFile(ctx, urlFor(webdavAddress1, path2), t)
	logrus.Info("6")
	if !ok {
		t.Error("error getting file")
		return
	}
	if fetchedContent != fileContents2 {
		t.Error("unexpected contents", fetchedContent)
		return
	}
	cancel()
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

	if resp.StatusCode >= 300 {
		t.Error("error response with status", resp.Status)
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

func urlFor(host string, path string) string {
	return "http://" + host + path
}
