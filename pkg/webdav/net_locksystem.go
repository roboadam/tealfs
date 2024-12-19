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

	"github.com/google/uuid"
	"golang.org/x/net/webdav"
)

type LockMessage interface {
	GetId() model.LockMessageId
}

type LockMessageReqResp struct {
	Req  LockMessage
	Resp chan LockMessage
}

type NetLockSystem struct {
	Messages chan LockMessageReqResp
	Release  chan model.LockMessageId
}

func NewNetLockSystem() *NetLockSystem {
	return &NetLockSystem{
		Messages: make(chan LockMessageReqResp),
	}
}

func (l *NetLockSystem) Confirm(now time.Time, name0 string, name1 string, conditions ...webdav.Condition) (release func(), err error) {
	req := model.LockConfirmRequest{
		Now:        now,
		Name0:      name0,
		Name1:      name1,
		Conditions: conditions,
		Id:         model.LockMessageId(uuid.New().String()),
	}

	respChan := make(chan LockMessage)

	l.Messages <- LockMessageReqResp{Req: &req, Resp: respChan}
	resp := <-respChan

	lcr, ok := resp.(*model.LockConfirmResponse)
	if !ok {
		return nil, errors.New("Not a confirm response")
	}

	if lcr.Ok {
		return func() {
			l.Release <- lcr.ReleaseId
		}, nil
	}
	return nil, errors.New(lcr.Message)
}

func (l *NetLockSystem) Create(now time.Time, details webdav.LockDetails) (token string, err error) {
	req := model.LockCreateRequest{
		Now:     now,
		Details: details,
	}
	respChan := make(chan LockMessage)

	l.Messages <- LockMessageReqResp{Req: &req, Resp: respChan}
	resp := <-respChan

	lcr, ok := resp.(*model.LockCreateResponse)
	if !ok {
		return "", errors.New("Not a create response")
	}

	if lcr.Ok {
		return string(lcr.Token), nil
	}
	return "", errors.New(lcr.Message)
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
