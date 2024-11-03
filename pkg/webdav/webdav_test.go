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
