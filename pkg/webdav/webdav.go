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
	"fmt"
	"tealfs/pkg/model"

	"golang.org/x/net/webdav"
)

type Webdav struct {
	webdavOps     WebdavOps
	webdavMgrGets chan model.ReadRequest
	webdavMgrPuts chan model.Block
	mgrWebdavGets chan model.ReadResult
	mgrWebdavPuts chan model.WriteResult
	fileSystem    FileSystem
	nodeId        model.NodeId
	pendingReads  map[model.BlockId]chan []byte
	pendingPuts   map[model.BlockId]chan error
}

func New(nodeId model.NodeId) Webdav {
	w := Webdav{
		webdavOps:    &HttpWebdavOps{},
		fileSystem:   NewFileSystem(),
		nodeId:       nodeId,
		pendingReads: make(map[model.BlockId]chan []byte),
	}
	w.start()
	return w
}

func (w *Webdav) start() {
	lockSystem := LockSystem{}
	go w.eventLoop()

	handler := &webdav.Handler{
		Prefix:     "/",
		FileSystem: &w.fileSystem,
		LockSystem: &lockSystem,
	}

	w.webdavOps.Handle("/", handler)
	w.webdavOps.ListenAndServe(":8080")
}

func (w *Webdav) eventLoop() {
	for {
		select {
		case r := <-w.mgrWebdavGets:
			ch, ok := w.pendingReads[r.Block.Id]
			if ok {
				ch <- r.Block.Data
				delete(w.pendingReads, r.Block.Id)
			}
		case r:= <-w.mgrWebdavPuts:
			ch, ok := w.pendingPuts[r.]
			if ok {
				ch <- r.Block.Data
				delete(w.pendingReads, r.Block.Id)
			}
		case r := <-w.fileSystem.FetchBlockReq:
			w.webdavMgrGets <- model.ReadRequest{
				Caller:  w.nodeId,
				BlockId: r.Id,
			}
			w.pendingReads[r.Id] = r.Resp
		case r := <-w.fileSystem.PushBlockReq:
			w.webdavMgrPuts <- model.Block{
				Id:   r.Id,
				Data: r.Data,
			}
			w.pendingPuts[r.Id] = r.Resp
		}
	}
}
