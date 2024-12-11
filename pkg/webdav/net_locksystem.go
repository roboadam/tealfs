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
	"errors"
	"tealfs/pkg/model"
	"time"

	"golang.org/x/net/webdav"
)

type NetLockSystem struct {
	ConfirmChan chan struct {
		Req  model.LockConfirmRequest
		Resp chan model.LockConfirmResponse
	}
	ReleaseChan chan model.LockReleaseId
	CreateChan  chan struct {
		Req  model.LockCreateRequest
		Resp chan model.LockCreateResponse
	}
	RefreshChan chan struct {
		Req  model.LockRefreshRequest
		Resp chan model.LockRefreshResponse
	}
	UnlockChan chan struct {
		Req  model.LockUnlockRequest
		Resp chan model.LockUnlockResponse
	}
}

func NewNetLockSystem() *NetLockSystem {
	return &NetLockSystem{
		ConfirmChan: make(chan struct {
			Req  model.LockConfirmRequest
			Resp chan model.LockConfirmResponse
		}),
		ReleaseChan: make(chan model.LockReleaseId),
		CreateChan: make(chan struct {
			Req  model.LockCreateRequest
			Resp chan model.LockCreateResponse
		}),
		RefreshChan: make(chan struct {
			Req  model.LockRefreshRequest
			Resp chan model.LockRefreshResponse
		}),
		UnlockChan: make(chan struct {
			Req  model.LockUnlockRequest
			Resp chan model.LockUnlockResponse
		}),
	}
}

func (l *NetLockSystem) Confirm(now time.Time, name0 string, name1 string, conditions ...webdav.Condition) (release func(), err error) {
	req := model.LockConfirmRequest{
		Now:        now,
		Name0:      name0,
		Name1:      name1,
		Conditions: conditions,
	}

	respChan := make(chan model.LockConfirmResponse)

	l.ConfirmChan <- struct {
		Req  model.LockConfirmRequest
		Resp chan model.LockConfirmResponse
	}{Req: req, Resp: respChan}

	resp := <-respChan

	if resp.Ok {
		return func() {
			l.ReleaseChan <- resp.ReleaseId
		}, nil
	}
	return nil, errors.New(resp.Message)
}

func (l *NetLockSystem) Create(now time.Time, details webdav.LockDetails) (token string, err error) {
	req := model.LockCreateRequest{
		Now:     now,
		Details: details,
	}
	respChan := make(chan model.LockCreateResponse)

	l.CreateChan <- struct {
		Req  model.LockCreateRequest
		Resp chan model.LockCreateResponse
	}{Req: req, Resp: respChan}

	resp := <-respChan
	if resp.Ok {
		return string(resp.Token), nil
	}
	return "", errors.New(resp.Message)
}

func (l *NetLockSystem) Refresh(now time.Time, token string, duration time.Duration) (webdav.LockDetails, error) {
	req := model.LockRefreshRequest{
		Now:      now,
		Token:    model.LockToken(token),
		Duration: duration,
	}
	respChan := make(chan model.LockRefreshResponse)

	l.RefreshChan <- struct {
		Req  model.LockRefreshRequest
		Resp chan model.LockRefreshResponse
	}{Req: req, Resp: respChan}

	resp := <-respChan
	if resp.Ok {
		return resp.Details, nil
	}
	return webdav.LockDetails{}, errors.New(resp.Message)
}

func (l *NetLockSystem) Unlock(now time.Time, token string) error {
	req := model.LockUnlockRequest{
		Now:   now,
		Token: model.LockToken(token),
	}
	respChan := make(chan model.LockUnlockResponse)

	l.UnlockChan <- struct {
		Req  model.LockUnlockRequest
		Resp chan model.LockUnlockResponse
	}{Req: req, Resp: respChan}

	resp := <-respChan
	if resp.Ok {
		return nil
	}
	return errors.New(resp.Message)
}
