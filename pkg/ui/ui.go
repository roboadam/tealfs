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

	"github.com/google/uuid"
)

type Ui struct {
	NodeConnMap *model.NodeConnectionMapper

	connToReq   chan model.ConnectToNodeReq
	addDiskMsg  chan model.AddDiskMsg
	addDiskResp chan model.UiDiskStatus

	diskStatuses map[model.DiskId]model.UiDiskStatus
	sMux         sync.Mutex
	ops          HtmlOps
	nodeId       model.NodeId
	ctx          context.Context
}

func NewUi(
	connToReq chan model.ConnectToNodeReq,
	addDiskReq chan model.AddDiskMsg,
	addDiskResp chan model.UiDiskStatus,
	ops HtmlOps,
	nodeId model.NodeId,
	bindAddr string,
	ctx context.Context,
) *Ui {
	diskStatuses := make(map[model.DiskId]model.UiDiskStatus)
	ui := Ui{
		connToReq:    connToReq,
		addDiskMsg:   addDiskReq,
		addDiskResp:  addDiskResp,
		diskStatuses: diskStatuses,
		ops:          ops,
		nodeId:       nodeId,
		ctx:          ctx,
	}
	ui.handleRoot()
	ui.start(bindAddr)
	return &ui
}

func (ui *Ui) start(bindAddr string) {
	go ui.handleMessages()
	go ui.ops.ListenAndServe(bindAddr)
}

func (ui *Ui) handleMessages() {
	for {
		select {
		case <-ui.ctx.Done():
			ui.ops.Shutdown()
			return
		case diskStatus := <-ui.addDiskResp:
			ui.saveDiskStatus(diskStatus)
		}
	}
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
		switch r.Method {
		case http.MethodGet:
			ui.connectToGet(w, tmpl)
		case http.MethodPut:
			hostAndPort := r.FormValue("hostAndPort")
			ui.connToReq <- model.ConnectToNodeReq{Address: hostAndPort}
			ui.connectionStatus(w, tmpl)
		default:
			http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
		}
	})
	ui.ops.HandleFunc("/add-disk", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			ui.addDiskGet(w, tmpl, ui.nodeId)
		case http.MethodPut:
			diskPath := r.FormValue("diskPath")
			node := r.FormValue("node")
			if node == "" {
				node = string(ui.nodeId)
			}
			req := model.AddDiskMsg{
				DiskId: model.DiskId(uuid.NewString()),
				Path:   diskPath,
				NodeId: model.NodeId(node),
			}
			chanutil.Send(ui.ctx, ui.addDiskMsg, req, "ui: add disk req")
			ui.connectionStatus(w, tmpl)
		default:
			http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
		}
	})
}
