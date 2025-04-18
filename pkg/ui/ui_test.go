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

package ui_test

import (
	"context"
	"net/http"
	"net/url"
	"strings"
	"tealfs/pkg/model"
	"tealfs/pkg/ui"
	"testing"
)

func TestListenAddress(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	_, _, _, ops := NewUi(ctx)
	if ops.BindAddr != "mockBindAddr:123" {
		t.Error("Didn't bind to mockBindAddr:123")
	}
}

func TestConnectTo(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	_, connToReq, _, ops := NewUi(ctx)
	mockResponseWriter := ui.MockResponseWriter{}
	request := http.Request{
		Method:   http.MethodPut,
		PostForm: make(url.Values),
	}
	request.PostForm.Add("hostAndPort", "abcdef")

	go ops.Handlers["/connect-to"](&mockResponseWriter, &request)
	reqToMgr := <-connToReq
	if reqToMgr.Address != "abcdef" {
		t.Error("Didn't send proper request to Mgr")
	}
}

func TestAddDiskGet(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	_, _, connToResp, ops := NewUi(ctx)
	mockResponseWriter := ui.MockResponseWriter{}
	request := http.Request{Method: http.MethodGet}
	nodeId1 := model.NewNodeId()
	nodeId2 := model.NewNodeId()

	connToResp <- model.UiConnectionStatus{
		Type:          model.Connected,
		RemoteAddress: "1234",
		Id:            nodeId1,
	}
	connToResp <- model.UiConnectionStatus{
		Type:          model.Connected,
		RemoteAddress: "5678",
		Id:            nodeId2,
	}

	waitForWrittenData(func() string {
		ops.Handlers["/add-disk"](&mockResponseWriter, &request)
		return mockResponseWriter.WrittenData
	}, []string{"local", string(nodeId1), string(nodeId2)})
}

func TestStatus(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	_, _, connToResp, ops := NewUi(ctx)
	mockResponseWriter := ui.MockResponseWriter{}
	request := http.Request{
		Method:   http.MethodGet,
		PostForm: make(url.Values),
	}
	request.PostForm.Add("hostAndPort", "abcdef")

	connToResp <- model.UiConnectionStatus{
		Type:          model.Connected,
		RemoteAddress: "1234",
		Id:            model.NewNodeId(),
	}
	connToResp <- model.UiConnectionStatus{
		Type:          model.NotConnected,
		RemoteAddress: "5678",
		Id:            model.NewNodeId(),
	}

	waitForWrittenData(func() string {
		ops.Handlers["/connection-status"](&mockResponseWriter, &request)
		return mockResponseWriter.WrittenData
	}, []string{"1234", "5678"})
}

func waitForWrittenData(handler func() string, values []string) {
	for {
		result := handler()
		foundAll := true
		for _, value := range values {
			if !strings.Contains(result, value) {
				foundAll = false
				break
			}
		}
		if foundAll {
			return
		}
	}
}

func NewUi(ctx context.Context) (*ui.Ui, chan model.UiMgrConnectTo, chan model.UiConnectionStatus, *ui.MockHtmlOps) {
	connToReq := make(chan model.UiMgrConnectTo)
	connToResp := make(chan model.UiConnectionStatus)
	diskAddReq := make(chan model.AddDiskReq)
	diskStatus := make(chan model.UiDiskStatus)
	ops := ui.MockHtmlOps{
		BindAddr: "mockBindAddr:123",
		Handlers: make(map[string]func(http.ResponseWriter, *http.Request)),
	}
	u := ui.NewUi(connToReq, connToResp, diskAddReq, diskStatus, &ops, "nodeId", "address", ctx)
	return u, connToReq, connToResp, &ops
}
