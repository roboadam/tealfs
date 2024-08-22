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
	smux       sync.Mutex
	ops        HtmlOps
}

func NewUi(connToReq chan model.UiMgrConnectTo, connToResp chan model.ConnectionStatus, ops HtmlOps) Ui {
	statuses := make(map[model.ConnId]model.ConnectionStatus)
	return Ui{
		connToReq:  connToReq,
		connToResp: connToResp,
		statuses:   statuses,
		ops:        ops,
	}
}

func (ui *Ui) Start() {
	ui.registerHttpHandlers()
	ui.handleRoot()
	go ui.handleMessages()
	err := ui.ops.ListenAndServe(":0")
	if err != nil {
		os.Exit(1)
	}
}

func (ui *Ui) handleMessages() {
	for status := range ui.connToResp {
		ui.smux.Lock()
		ui.statuses[status.Id] = status
		ui.smux.Unlock()
	}
}

func (ui *Ui) registerHttpHandlers() {
	ui.ops.HandleFunc("/connect-to", func(w http.ResponseWriter, r *http.Request) {
		hostAndPort := r.FormValue("hostandport")
		ui.connToReq <- model.UiMgrConnectTo{Address: hostAndPort}
	})
}

func (ui *Ui) htmlStatus(divId string) string {
	var builder strings.Builder
	builder.WriteString(`<div id="`)
	builder.WriteString(divId)
	builder.WriteString(`">`)
	ui.smux.Lock()
	for _, value := range ui.statuses {
		builder.WriteString(string(value.Id))
		builder.WriteString(" ")
		builder.WriteString(string(value.Type))
		builder.WriteString("<br />")
	}
	ui.smux.Unlock()
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
					` + htmlMyhost("TODO") + `
					<p>Input the host and port of a node to add</p>
					<form hx-put="/connect-to">
						<label for="textbox">Host and port:</label>
						<input type="text" id="hostandport" name="hostandport">
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

func htmlMyhost(address string) string {
	return wrapInDiv(`
			<h2>My host</h2>
			<p>Host: `+address+`</p>`,
		"myhost")
}

func wrapInDiv(html string, divId string) string {
	return `<div id="` + divId + `">` + html + `</div>`
}
