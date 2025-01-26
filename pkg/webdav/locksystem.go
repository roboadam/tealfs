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

type LockSystemMode int8

const (
	LockSystemModeNet LockSystemMode = iota
	LockSystemModeLocal
)

type LockSystem struct {
	mode        LockSystemMode
	netLs       *NetLockSystem
	localLs     webdav.LockSystem
	MessageChan chan LockMessageReqResp
	ReleaseChan chan model.LockMessageId
}

func NewLockSystem(nodeId model.NodeId) *LockSystem {
	mode := LockSystemModeLocal
	netLs := NewNetLockSystem(nodeId)
	localLs := webdav.NewMemLS()
	return &LockSystem{
		mode:        mode,
		netLs:       netLs,
		localLs:     localLs,
		MessageChan: netLs.Messages,
		ReleaseChan: netLs.Release,
	}
}

func (l *LockSystem) UseNetLockSystem() {
	l.mode = LockSystemModeNet
}

func (l *LockSystem) UseLocalLockSystem() {
	l.mode = LockSystemModeLocal
}

func (l *LockSystem) Confirm(now time.Time, name0 string, name1 string, conditions ...webdav.Condition) (release func(), err error) {
	switch l.mode {
	case LockSystemModeLocal:
		return l.localLs.Confirm(now, name0, name1, conditions...)
	case LockSystemModeNet:
		f, e := l.netLs.Confirm(now, name0, name1, conditions...)
		return f, e
	default:
		return nil, errors.New("invalid lock mode")
	}
}

func (l *LockSystem) Create(now time.Time, details webdav.LockDetails) (token string, err error) {
	switch l.mode {
	case LockSystemModeLocal:
		return l.localLs.Create(now, details)
	case LockSystemModeNet:
		t, e := l.netLs.Create(now, details)
		return t, e
	default:
		return "", errors.New("invalid lock mode")
	}
}

func (l *LockSystem) Refresh(now time.Time, token string, duration time.Duration) (webdav.LockDetails, error) {
	switch l.mode {
	case LockSystemModeLocal:
		return l.localLs.Refresh(now, token, duration)
	case LockSystemModeNet:
		l, e := l.netLs.Refresh(now, token, duration)
		return l, e
	default:
		return webdav.LockDetails{}, errors.New("invalid lock mode")
	}
}

func (l *LockSystem) Unlock(now time.Time, token string) error {
	switch l.mode {
	case LockSystemModeLocal:
		return l.localLs.Unlock(now, token)
	case LockSystemModeNet:
		e := l.netLs.Unlock(now, token)
		return e
	default:
		return errors.New("invalid lock mode")
	}
}
