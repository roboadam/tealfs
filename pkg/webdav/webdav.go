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

package webdav

import (
	"context"
	"net/http"
	"tealfs/pkg/chanutil"
	"tealfs/pkg/disk"
	"tealfs/pkg/model"

	log "github.com/sirupsen/logrus"

	"golang.org/x/net/webdav"
)

type Webdav struct {
	webdavMgrGets chan model.GetBlockReq
	webdavMgrPuts chan model.PutBlockReq
	mgrWebdavGets chan model.GetBlockResp
	mgrWebdavPuts chan model.PutBlockResp

	fileSystem   FileSystem
	nodeId       model.NodeId
	pendingReads map[model.GetBlockId]chan model.GetBlockResp
	pendingPuts  map[model.PutBlockId]chan model.PutBlockResp
	lockSystem   webdav.LockSystem
	bindAddress  string
	server       *http.Server

	ctx context.Context
}

func New(
	nodeId model.NodeId,
	webdavMgrGets chan model.GetBlockReq,
	webdavMgrPuts chan model.PutBlockReq,
	mgrWebdavGets chan model.GetBlockResp,
	mgrWebdavPuts chan model.PutBlockResp,
	outSends chan model.MgrConnsSend,

	mgrWebdavBroadcast chan FileBroadcast,
	bindAddress string,
	ctx context.Context,
	fileOps disk.FileOps,
	indexPath string,
	chansize int,
	mapper *model.NodeConnectionMapper,
) Webdav {
	w := Webdav{
		webdavMgrGets: webdavMgrGets,
		webdavMgrPuts: webdavMgrPuts,
		mgrWebdavGets: mgrWebdavGets,
		mgrWebdavPuts: mgrWebdavPuts,
		fileSystem:    NewFileSystem(nodeId, mgrWebdavBroadcast, fileOps, indexPath, chansize, outSends, mapper, ctx),
		nodeId:        nodeId,
		pendingReads:  make(map[model.GetBlockId]chan model.GetBlockResp),
		pendingPuts:   make(map[model.PutBlockId]chan model.PutBlockResp),
		lockSystem:    webdav.NewMemLS(),
		bindAddress:   bindAddress,
		ctx:           ctx,
	}
	w.fileSystem.OutSends = outSends
	w.fileSystem.Mapper = mapper
	w.start()
	return w
}

func (w *Webdav) start() {
	go w.eventLoop()

	handler := &webdav.Handler{
		Prefix:     "/",
		FileSystem: &w.fileSystem,
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
		case r := <-w.fileSystem.ReadReqResp:
			chanutil.Send(w.ctx, w.webdavMgrGets, r.Req, "webdav: read request to mgr "+string(r.Req.Id))
			w.pendingReads[r.Req.Id] = r.Resp
		case r := <-w.fileSystem.WriteReqResp:
			chanutil.Send(w.ctx, w.webdavMgrPuts, r.Req, "webdav: write request to mgr")
			w.pendingPuts[r.Req.Id] = r.Resp
		}
	}
}
