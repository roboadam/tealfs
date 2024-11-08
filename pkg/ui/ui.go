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

package ui

import (
	"fmt"
	"net/http"
	"os"
	"strings"
	"sync"
	"tealfs/pkg/model"
)

type Ui struct {
	connToReq  chan model.UiMgrConnectTo
	connToResp chan model.ConnectionStatus
	statuses   map[model.ConnId]model.ConnectionStatus
	sMux       sync.Mutex
	ops        HtmlOps
}

func NewUi(connToReq chan model.UiMgrConnectTo, connToResp chan model.ConnectionStatus, ops HtmlOps) *Ui {
	statuses := make(map[model.ConnId]model.ConnectionStatus)
	ui := Ui{
		connToReq:  connToReq,
		connToResp: connToResp,
		statuses:   statuses,
		ops:        ops,
	}
	ui.registerHttpHandlers()
	ui.handleRoot()
	go ui.start()
	return &ui
}

func (ui *Ui) start() {
	go ui.handleMessages()
	err := ui.ops.ListenAndServe("localhost:8081")
	if err != nil {
		os.Exit(1)
	}
}

func (ui *Ui) handleMessages() {
	for status := range ui.connToResp {
		ui.saveStatus(status)
	}
}

func (ui *Ui) saveStatus(status model.ConnectionStatus) {
	ui.sMux.Lock()
	defer ui.sMux.Unlock()
	ui.statuses[status.Id] = status
}

func (ui *Ui) registerHttpHandlers() {
	ui.ops.HandleFunc("/connect-to", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
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
		html := `
			<!DOCTYPE html>
			<html>
			<head>
				<title>TealFS</title>
				<link rel="stylesheet" href="https://unpkg.com/mvp.css@1.12/mvp.css" /> 
				<script src="https://unpkg.com/htmx.org@1.9.2"></script>
			</head>
			<body>
			    <main>
					<h1>TealFS</h1>
					<p>Input the host and port of a node to add</p>
					<form hx-put="/connect-to">
						<label for="textbox">Host and port:</label>
						<input type="text" id="hostAndPort" name="hostAndPort">
						<input type="submit" value="Connect">
					</form>
					` + ui.htmlStatus("status") + `
				</main>
			</body>
			</html>
		`

		// Write the HTML content to the response writer
		_, _ = fmt.Fprintf(w, "%s", html)
	})
}
