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

package webdav

import (
	"context"
	"net/http"
	"tealfs/pkg/model"

	"golang.org/x/net/webdav"
)

type Webdav struct {
	webdavMgrGets chan model.ReadRequest
	webdavMgrPuts chan model.WriteRequest
	mgrWebdavGets chan model.ReadResult
	mgrWebdavPuts chan model.WriteResult
	fileSystem    FileSystem
	nodeId        model.NodeId
	pendingReads  map[model.BlockId]chan model.ReadResult
	pendingPuts   map[model.BlockId]chan model.WriteResult
	bindAddress   string
	server        http.Server
}

func New(
	nodeId model.NodeId,
	webdavMgrGets chan model.ReadRequest,
	webdavMgrPuts chan model.WriteRequest,
	mgrWebdavGets chan model.ReadResult,
	mgrWebdavPuts chan model.WriteResult,
	bindAddress string,
	ctx context.Context,
) Webdav {
	w := Webdav{
		webdavMgrGets: webdavMgrGets,
		webdavMgrPuts: webdavMgrPuts,
		mgrWebdavGets: mgrWebdavGets,
		mgrWebdavPuts: mgrWebdavPuts,
		fileSystem:    NewFileSystem(nodeId),
		nodeId:        nodeId,
		pendingReads:  make(map[model.BlockId]chan model.ReadResult),
		pendingPuts:   make(map[model.BlockId]chan model.WriteResult),
		bindAddress:   bindAddress,
	}
	w.start(ctx)
	return w
}

func (w *Webdav) StopWebdavServer() {
}

func (w *Webdav) start(ctx context.Context) {
	lockSystem := LockSystem{
		locks: make(map[string]webdav.LockDetails),
	}
	go w.eventLoop(ctx)

	handler := &webdav.Handler{
		Prefix:     "/",
		FileSystem: &w.fileSystem,
		LockSystem: &lockSystem,
	}

	mux := http.NewServeMux()
	mux.Handle("/", handler)
	w.server = http.Server{
		Addr:    w.bindAddress,
		Handler: mux,
	}
	go func() {
		err := w.server.ListenAndServe()
		if err != nil {
			
		}
	}()
}

func (w *Webdav) eventLoop(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			w.server.Shutdown(ctx)
		case r := <-w.mgrWebdavGets:
			ch, ok := w.pendingReads[r.Block.Id]
			if ok {
				ch <- r
				delete(w.pendingReads, r.Block.Id)
			}
		case r := <-w.mgrWebdavPuts:
			ch, ok := w.pendingPuts[r.BlockId]
			if ok {
				ch <- r
				delete(w.pendingPuts, r.BlockId)
			}
		case r := <-w.fileSystem.ReadReqResp:
			w.webdavMgrGets <- r.Req
			w.pendingReads[r.Req.BlockId] = r.Resp
		case r := <-w.fileSystem.WriteReqResp:
			w.webdavMgrPuts <- r.Req
			w.pendingPuts[r.Req.Block.Id] = r.Resp
		}
	}
}
