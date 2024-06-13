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
	"tealfs/pkg/mgr"
)

type Ui struct {
	connToReq  chan mgr.UiMgrConnectTo
	connToResp chan mgr.ConnectToResp
}

func NewUi(connToReq chan mgr.UiMgrConnectTo) Ui {
	connToResp := make(chan mgr.ConnectToResp, 100)
	return Ui{connToReq, connToResp}
}

func (ui Ui) Start() {
	ui.registerHttpHandlers()
	ui.handleRoot()
	err := http.ListenAndServe(":0", nil)
	if err != nil {
		os.Exit(1)
	}
}

func (ui Ui) registerHttpHandlers() {
	http.HandleFunc("/connect-to", func(w http.ResponseWriter, r *http.Request) {
		hostAndPort := r.FormValue("hostandport")
		ui.connToReq <- mgr.UiMgrConnectTo{Address: hostAndPort}
		_, _ = fmt.Fprintf(w, "Connecting to: %s", hostAndPort)
		resp := <-ui.connToResp
		if resp.Success {
			_, _ = fmt.Fprintf(w, "Connected! to: %s", string(resp.Id))
		} else {
			_, _ = fmt.Fprintf(w, "Connection Failure: %s", resp.ErrorMessage)
		}
	})
}

func (ui Ui) handleRoot() {
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
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
