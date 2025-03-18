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

package ui

import (
	"context"
	"fmt"
	"net/http"
	"strings"
	"sync"
	"tealfs/pkg/model"
)

type Ui struct {
	connToReq  chan model.UiMgrConnectTo
	connToResp chan model.UiConnectionStatus
	statuses   map[model.NodeId]model.UiConnectionStatus
	sMux       sync.Mutex
	ops        HtmlOps
}

func NewUi(connToReq chan model.UiMgrConnectTo, connToResp chan model.UiConnectionStatus, ops HtmlOps, bindAddr string, ctx context.Context) *Ui {
	statuses := make(map[model.NodeId]model.UiConnectionStatus)
	ui := Ui{
		connToReq:  connToReq,
		connToResp: connToResp,
		statuses:   statuses,
		ops:        ops,
	}
	ui.registerHttpHandlers()
	ui.handleRoot()
	ui.start(bindAddr, ctx)
	return &ui
}

func (ui *Ui) start(bindAddr string, ctx context.Context) {
	go ui.handleMessages(ctx)
	go ui.ops.ListenAndServe(bindAddr)
}

func (ui *Ui) handleMessages(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			ui.ops.Shutdown()
			return
		case status := <-ui.connToResp:
			ui.saveStatus(status)
		}
	}
}

func (ui *Ui) saveStatus(status model.UiConnectionStatus) {
	ui.sMux.Lock()
	defer ui.sMux.Unlock()
	ui.statuses[status.Id] = status
}

func (ui *Ui) registerHttpHandlers() {
	ui.ops.HandleFunc("/connect-to", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPut {
			http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
			return
		}
		hostAndPort := r.FormValue("hostAndPort")
		ui.connToReq <- model.UiMgrConnectTo{Address: hostAndPort}
	})
}

func (ui *Ui) htmlStatus(divId string) string {
	var builder strings.Builder
	builder.WriteString(`<div id="`)
	builder.WriteString(divId)
	builder.WriteString(`">`)
	ui.sMux.Lock()
	for _, value := range ui.statuses {
		builder.WriteString(string(value.RemoteAddress))
		builder.WriteString(" ")
		builder.WriteString(fmt.Sprint(value.Type))
		builder.WriteString("<br />")
	}
	ui.sMux.Unlock()
	builder.WriteString("</div>")
	return builder.String()
}

func (ui *Ui) handleRoot() {
	ui.ops.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		templ(w, ui.htmlStatus("status"))
	})
}
