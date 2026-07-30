// Copyright (C) 2026 Adam Hess
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
	"encoding/gob"
	"net/http"
	"tealfs/pkg/chanutil"
	"tealfs/pkg/disk"
	"tealfs/pkg/model"

	log "github.com/sirupsen/logrus"

	"golang.org/x/net/webdav"
)

func init() {
	gob.Register(&FileBroadcast{})
}

type Webdav struct {
	webdavMgrPuts chan model.PutBlockReq
	mgrWebdavGets chan model.FetchBlockResp
	mgrWebdavPuts chan model.PutBlockResp
	outPayloads   chan model.Payload2

	FileSystem   FileSystem
	nodeId       model.NodeId
	pendingReads map[model.FetchBlockId]chan model.FetchBlockResp
	pendingPuts  map[model.PutBlockId]chan model.PutBlockResp
	lockSystem   webdav.LockSystem
	bindAddress  string
	server       *http.Server

	ctx context.Context
}

func New(
	nodeId model.NodeId,
	webdavMgrPuts chan model.PutBlockReq,
	mgrWebdavGets chan model.FetchBlockResp,
	mgrWebdavPuts chan model.PutBlockResp,
	outPayloads chan model.Payload2,

	mgrWebdavBroadcast chan FileBroadcast,
	bindAddress string,
	ctx context.Context,
	fileOps disk.FileOps,
	indexPath string,
	chansize int,
	mapper *model.NodeConnectionMapper,
) Webdav {
	w := Webdav{
		webdavMgrPuts: webdavMgrPuts,
		mgrWebdavGets: mgrWebdavGets,
		mgrWebdavPuts: mgrWebdavPuts,
		outPayloads:   outPayloads,
		FileSystem:    NewFileSystem(nodeId, mgrWebdavBroadcast, fileOps, indexPath, chansize, mapper, ctx),
		nodeId:        nodeId,
		pendingReads:  make(map[model.FetchBlockId]chan model.FetchBlockResp),
		pendingPuts:   make(map[model.PutBlockId]chan model.PutBlockResp),
		lockSystem:    webdav.NewMemLS(),
		bindAddress:   bindAddress,
		ctx:           ctx,
	}
	w.FileSystem.Mapper = mapper
	w.FileSystem.OutPayload = outPayloads
	w.start()
	return w
}

func (w *Webdav) start() {
	go w.eventLoop()

	handler := &webdav.Handler{
		Prefix:     "/",
		FileSystem: &w.FileSystem,
		LockSystem: w.lockSystem,
	}

	mux := http.NewServeMux()
	mux.Handle("/", handler)
	w.server = &http.Server{
		Addr:    w.bindAddress,
		Handler: mux,
	}
	go w.server.ListenAndServe()
}

func (w *Webdav) eventLoop() {
	for {
		select {
		case <-w.ctx.Done():
			w.server.Close()
			return
		case r := <-w.mgrWebdavGets:
			ch, ok := w.pendingReads[r.Id]
			if ok {
				chanutil.Send(w.ctx, ch, r, "webdav: response for pending read to fs")
				delete(w.pendingReads, r.Id)
			}
		case r := <-w.mgrWebdavPuts:
			ch, ok := w.pendingPuts[r.Id]
			if ok {
				chanutil.Send(w.ctx, ch, r, "webdav: response for pending write to fs")
				delete(w.pendingPuts, r.Id)
			} else {
				log.Warn("webdav: received write response for unknown put block id", r.Id)
			}
		case r := <-w.FileSystem.ReadReqResp:
			w.outPayloads <- &r.Req
			w.pendingReads[r.Req.Id] = r.Resp
		case r := <-w.FileSystem.WriteReqResp:
			chanutil.Send(w.ctx, w.webdavMgrPuts, r.Req, "webdav: write request to mgr")
			w.pendingPuts[r.Req.Id] = r.Resp
		}
	}
}
