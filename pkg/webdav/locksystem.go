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
}

func NewLockSystem() *LockSystem {
	mode := LockSystemModeLocal
	netLs := NewNetLockSystem()
	localLs := webdav.NewMemLS()
	return &LockSystem{
		mode:        mode,
		netLs:       netLs,
		localLs:     localLs,
		MessageChan: netLs.Messages,
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
		return l.netLs.Confirm(now, name0, name1, conditions...)
	default:
		return nil, errors.New("invalid lock mode")
	}
}

func (l *LockSystem) Create(now time.Time, details webdav.LockDetails) (token string, err error) {
	switch l.mode {
	case LockSystemModeLocal:
		return l.localLs.Create(now, details)
	case LockSystemModeNet:
		return l.netLs.Create(now, details)
	default:
		return "", errors.New("invalid lock mode")
	}
}

func (l *LockSystem) Refresh(now time.Time, token string, duration time.Duration) (webdav.LockDetails, error) {
	switch l.mode {
	case LockSystemModeLocal:
		return l.localLs.Refresh(now, token, duration)
	case LockSystemModeNet:
		return l.netLs.Refresh(now, token, duration)
	default:
		return webdav.LockDetails{}, errors.New("invalid lock mode")
	}
}

func (l *LockSystem) Unlock(now time.Time, token string) error {
	switch l.mode {
	case LockSystemModeLocal:
		return l.localLs.Unlock(now, token)
	case LockSystemModeNet:
		return l.netLs.Unlock(now, token)
	default:
		return errors.New("invalid lock mode")
	}
}
