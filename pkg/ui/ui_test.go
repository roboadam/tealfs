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

package ui_test

import (
	"net/http"
	"net/url"
	"strings"
	"tealfs/pkg/model"
	"tealfs/pkg/ui"
	"testing"
)

func TestListenAddress(t *testing.T) {
	_, _, _, ops := NewUi()
	if ops.BindAddr != ":0" {
		t.Error("Didn't bind to :0")
	}
}

func TestConnectTo(t *testing.T) {
	_, connToReq, _, ops := NewUi()
	mockResponseWriter := ui.MockResponseWriter{}
	request := http.Request{
		Method:   http.MethodPost,
		PostForm: make(url.Values),
	}
	request.PostForm.Add("hostAndPort", "abcdef")

	go ops.Handlers["/connect-to"](&mockResponseWriter, &request)
	reqToMgr := <-connToReq
	if reqToMgr.Address != "abcdef" {
		t.Error("Didn't send proper request to Mgr")
	}
}

func TestStatus(t *testing.T) {
	_, _, connToResp, ops := NewUi()
	mockResponseWriter := ui.MockResponseWriter{}
	request := http.Request{
		Method:   http.MethodGet,
		PostForm: make(url.Values),
	}
	request.PostForm.Add("hostAndPort", "abcdef")

	connToResp <- model.ConnectionStatus{
		Type:          model.Connected,
		RemoteAddress: "1234",
		Id:            model.ConnId(1),
	}
	connToResp <- model.ConnectionStatus{
		Type:          model.NotConnected,
		RemoteAddress: "5678",
		Id:            model.ConnId(2),
	}

	waitForWrittenData(func() string {
		ops.Handlers["/"](&mockResponseWriter, &request)
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

func NewUi() (*ui.Ui, chan model.UiMgrConnectTo, chan model.ConnectionStatus, *ui.MockHtmlOps) {
	connToReq := make(chan model.UiMgrConnectTo)
	connToResp := make(chan model.ConnectionStatus)
	ops := ui.MockHtmlOps{
		BindAddr: "",
		Handlers: make(map[string]func(http.ResponseWriter, *http.Request)),
	}
	u := ui.NewUi(connToReq, connToResp, &ops)
	u.Start()
	return &u, connToReq, connToResp, &ops
}
