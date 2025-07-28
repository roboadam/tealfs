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
	"tealfs/pkg/set"

	"github.com/google/uuid"
)

type Ui struct {
	connToReq   chan model.ConnectToNodeReq
	connToResp  chan model.UiConnectionStatus
	addDiskReq  chan model.AddDiskReq
	addDiskResp chan model.UiDiskStatus

	statuses     map[model.NodeId]model.UiConnectionStatus
	diskStatuses map[model.DiskId]model.UiDiskStatus
	remotes      set.Set[model.NodeId]
	sMux         sync.Mutex
	ops          HtmlOps
	nodeId       model.NodeId
	ctx          context.Context
}

func NewUi(
	connToReq chan model.ConnectToNodeReq,
	connToResp chan model.UiConnectionStatus,
	addDiskReq chan model.AddDiskReq,
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
		remotes:      set.NewSet[model.NodeId](),
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
	if status.Type == model.Connected {
		ui.remotes.Add(status.Id)
	} else {
		ui.remotes.Remove(status.Id)
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
			ui.addDiskGet(w, tmpl, ui.remotes.GetValues(), ui.nodeId)
		case http.MethodPut:
			diskPath := r.FormValue("diskPath")
			node := r.FormValue("node")
			if node == "" {
				node = string(ui.nodeId)
			}
			req := model.AddDiskReq{
				Id:   model.DiskId(uuid.NewString()),
				Path: diskPath,
				Node: model.NodeId(node),
			}
			chanutil.Send(ui.ctx, ui.addDiskReq, req, "ui: add disk req")
			ui.connectionStatus(w, tmpl)
		default:
			http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
		}
	})
}
