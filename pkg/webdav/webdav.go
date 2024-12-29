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
	webdavMgrGets      chan model.ReadRequest
	webdavMgrPuts      chan model.WriteRequest
	mgrWebdavGets      chan model.ReadResult
	mgrWebdavPuts      chan model.WriteResult
	mgrWebdavIsPrimary chan bool
	webdavMgrLockMsg   chan LockMessage
	mgrWebdavLockMsg   chan LockMessage
	fileSystem         FileSystem
	nodeId             model.NodeId
	pendingReads       map[model.BlockId]chan model.ReadResult
	pendingPuts        map[model.BlockId]chan model.WriteResult
	pendingLockMsg     map[model.LockMessageId]chan LockMessage
	pendingReleases    map[model.LockMessageId]func()
	lockSystem         *LockSystem
	bindAddress        string
	server             *http.Server
}

func New(
	nodeId model.NodeId,
	webdavMgrGets chan model.ReadRequest,
	webdavMgrPuts chan model.WriteRequest,
	mgrWebdavGets chan model.ReadResult,
	mgrWebdavPuts chan model.WriteResult,
	mgrWebdavIsPrimary chan bool,
	webdavMgrLockMsg chan LockMessage,
	mgrWebdavLockMsg chan LockMessage,
	bindAddress string,
	ctx context.Context,
) Webdav {
	w := Webdav{
		webdavMgrGets:      webdavMgrGets,
		webdavMgrPuts:      webdavMgrPuts,
		mgrWebdavGets:      mgrWebdavGets,
		mgrWebdavPuts:      mgrWebdavPuts,
		mgrWebdavIsPrimary: mgrWebdavIsPrimary,
		webdavMgrLockMsg:   webdavMgrLockMsg,
		mgrWebdavLockMsg:   mgrWebdavLockMsg,
		fileSystem:         NewFileSystem(nodeId),
		nodeId:             nodeId,
		pendingReads:       make(map[model.BlockId]chan model.ReadResult),
		pendingPuts:        make(map[model.BlockId]chan model.WriteResult),
		pendingLockMsg:     make(map[model.LockMessageId]chan LockMessage),
		pendingReleases:    map[model.LockMessageId]func(){},
		lockSystem:         NewLockSystem(nodeId),
		bindAddress:        bindAddress,
	}
	w.start(ctx)
	return w
}

func (w *Webdav) start(ctx context.Context) {
	go w.eventLoop(ctx)

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

func (w *Webdav) eventLoop(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			w.server.Shutdown(context.Background())
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
		case r := <-w.lockSystem.MessageChan:
			w.webdavMgrLockMsg <- r.Req
			w.pendingLockMsg[r.Req.GetId()] = r.Resp
		case r := <-w.lockSystem.ReleaseChan:
			w.webdavMgrLockMsg <- r
		case r := <-w.mgrWebdavLockMsg:
			ch, ok := w.pendingLockMsg[r.GetId()]
			if ok {
				ch <- r
				delete(w.pendingLockMsg, r.GetId())
			} else {
				switch msg := r.(type) {
				case *model.LockConfirmRequest:
					release, err := w.lockSystem.Confirm(msg.Now, msg.Name0, msg.Name1, msg.Conditions...)
					if err != nil {
						w.webdavMgrLockMsg <- &model.LockConfirmResponse{
							Ok:      false,
							Id:      msg.Id,
							Message: err.Error(),
							Caller:  msg.Caller,
						}
					} else {
						w.pendingReleases[msg.Id] = release
						w.webdavMgrLockMsg <- &model.LockConfirmResponse{
							Ok:     true,
							Id:     msg.Id,
							Caller: msg.Caller,
						}
					}
				case *model.LockMessageId:
					release, ok := w.pendingReleases[*msg]
					if ok {
						release()
						delete(w.pendingReleases, *msg)
					}
				case *model.LockUnlockRequest:
					err := w.lockSystem.Unlock(msg.Now, string(msg.Token))
					if err == nil {
						w.webdavMgrLockMsg <- &model.LockUnlockResponse{
							Ok:     true,
							Id:     msg.Id,
							Caller: msg.Caller,
						}
					} else {
						w.webdavMgrLockMsg <- &model.LockUnlockResponse{
							Ok:      false,
							Message: err.Error(),
							Id:      msg.Id,
							Caller:  msg.Caller,
						}
					}
				case *model.LockCreateRequest:
					token, err := w.lockSystem.Create(msg.Now, msg.Details)
					if err == nil {
						w.webdavMgrLockMsg <- &model.LockCreateResponse{
							Ok:     true,
							Token:  model.LockToken(token),
							Id:     msg.Id,
							Caller: msg.Caller,
						}
					} else {
						w.webdavMgrLockMsg <- &model.LockCreateResponse{
							Ok:      false,
							Message: err.Error(),
							Id:      msg.Id,
							Caller:  msg.Caller,
						}
					}
				case *model.LockRefreshRequest:
					details, err := w.lockSystem.Refresh(msg.Now, string(msg.Token), msg.Duration)
					if err == nil {
						w.webdavMgrLockMsg <- &model.LockRefreshResponse{
							Ok:      true,
							Details: details,
							Id:      msg.Id,
							Caller:  msg.Caller,
						}
					} else {
						w.webdavMgrLockMsg <- &model.LockRefreshResponse{
							Ok:      false,
							Message: err.Error(),
							Id:      msg.Id,
							Caller:  msg.Caller,
						}
					}
				}
			}
		case r := <-w.fileSystem.ReadReqResp:
			w.webdavMgrGets <- r.Req
			w.pendingReads[r.Req.BlockId] = r.Resp
		case r := <-w.fileSystem.WriteReqResp:
			w.webdavMgrPuts <- r.Req
			w.pendingPuts[r.Req.Block.Id] = r.Resp
		case isPrimary := <-w.mgrWebdavIsPrimary:
			if isPrimary {
				w.lockSystem.UseLocalLockSystem()
			} else {
				w.lockSystem.UseNetLockSystem()
			}
		}
	}
}
