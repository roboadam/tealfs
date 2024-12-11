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
	confirm chan struct {
		req  model.LockConfirmRequest
		resp chan model.LockConfirmResponse
	}
	release chan model.LockReleaseId
	create  chan struct {
		req  model.LockCreateRequest
		resp chan model.LockCreateResponse
	}
	refresh chan struct {
		req  model.LockRefreshRequest
		resp chan model.LockRefreshResponse
	}
	unlock chan struct {
		req  model.LockUnlockRequest
		resp chan model.LockUnlockResponse
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

	l.confirm <- struct {
		req  model.LockConfirmRequest
		resp chan model.LockConfirmResponse
	}{req: req, resp: respChan}

	resp := <-respChan

	if resp.Ok {
		return func() {
			l.release <- resp.ReleaseId
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

	l.create <- struct {
		req  model.LockCreateRequest
		resp chan model.LockCreateResponse
	}{req: req, resp: respChan}

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

	l.refresh <- struct {
		req  model.LockRefreshRequest
		resp chan model.LockRefreshResponse
	}{req: req, resp: respChan}

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

	l.unlock <- struct {
		req  model.LockUnlockRequest
		resp chan model.LockUnlockResponse
	}{req: req, resp: respChan}

	resp := <-respChan
	if resp.Ok {
		return nil
	}
	return errors.New(resp.Message)
}
