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
	"net/http"
	"sync"
	"tealfs/pkg/chanutil"
	"tealfs/pkg/model"
)

type Ui struct {
	connToReq   chan model.UiMgrConnectTo
	connToResp  chan model.UiConnectionStatus
	addDiskReq  chan model.DiskInfo
	addDiskResp chan model.UiDiskStatus

	statuses     map[model.NodeId]model.UiConnectionStatus
	diskStatuses map[model.DiskId]model.UiDiskStatus
	sMux         sync.Mutex
	ops          HtmlOps
	nodeId       model.NodeId
}

func NewUi(
	connToReq chan model.UiMgrConnectTo,
	connToResp chan model.UiConnectionStatus,
	addDiskReq chan model.DiskInfo,
	addDiskResp chan model.UiDiskStatus,
	ops HtmlOps,
	nodeId model.NodeId,
	bindAddr string,
	ctx context.Context,
) *Ui {
	statuses := make(map[model.NodeId]model.UiConnectionStatus)
	diskStatuses := make(map[model.DiskId]model.UiDiskStatus)
	ui := Ui{
		connToReq:    connToReq,
		connToResp:   connToResp,
		addDiskReq:   addDiskReq,
		addDiskResp:  addDiskResp,
		statuses:     statuses,
		diskStatuses: diskStatuses,
		ops:          ops,
		nodeId:       nodeId,
	}
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
		case diskStatus := <-ui.addDiskResp:
			ui.saveDiskStatus(diskStatus)
		}
	}
}

func (ui *Ui) saveStatus(status model.UiConnectionStatus) {
	ui.sMux.Lock()
	defer ui.sMux.Unlock()
	ui.statuses[status.Id] = status
}

func (ui *Ui) saveDiskStatus(status model.UiDiskStatus) {
	ui.sMux.Lock()
	defer ui.sMux.Unlock()
	ui.diskStatuses[status.Id] = status
}

func (ui *Ui) handleRoot() {
	tmpl := initTemplates()
	ui.ops.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		ui.index(w, tmpl)
	})
	ui.ops.HandleFunc("/connection-status", func(w http.ResponseWriter, r *http.Request) {
		ui.connectionStatus(w, tmpl)
	})
	ui.ops.HandleFunc("/disk-status", func(w http.ResponseWriter, r *http.Request) {
		ui.diskStatus(w, tmpl)
	})
	ui.ops.HandleFunc("/connect-to", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet {
			ui.connectToGet(w, tmpl)
		} else if r.Method == http.MethodPut {
			hostAndPort := r.FormValue("hostAndPort")
			ui.connToReq <- model.UiMgrConnectTo{Address: hostAndPort}
			ui.connectionStatus(w, tmpl)
		} else {
			http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
		}
	})
	ui.ops.HandleFunc("/add-disk", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet {
			ui.addDiskGet(w, tmpl)
		} else if r.Method == http.MethodPut {
			diskPath := r.FormValue("diskPath")
			req := model.NewDiskInfo(diskPath, ui.nodeId, 1)
			chanutil.Send(ui.addDiskReq, req, "ui: add disk req")
			ui.connectionStatus(w, tmpl)
		} else {
			http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
		}
	})
}
