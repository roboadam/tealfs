package ui_test

import (
	"net/http"
	"tealfs/pkg/model"
	"tealfs/pkg/ui"
	"testing"
)

func TestListenAddress(t *testing.T) {
	connToReq := make(chan model.UiMgrConnectTo)
	connToResp := make(chan model.ConnectionStatus)
	ops := ui.MockHtmlOps{
		BindAddr: "",
		Handlers: make(map[string]func(http.ResponseWriter, *http.Request)),
	}
	u := ui.NewUi(connToReq, connToResp, &ops)
	u.Start()
	if ops.BindAddr != ":0" {
		t.Error("Didn't bind to :0")
	}
}
