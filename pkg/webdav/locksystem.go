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
	"time"

	"github.com/google/uuid"
	"golang.org/x/net/webdav"
)

type LockSystem struct {
	value webdav.LockSystem
}

func (l *LockSystem) UseNetLockSystem() {
	l.value = NewNetLockSystem()
}

func (l *LockSystem) UseLocalLockSystem() {
	l.value = webdav.NewMemLS()
}

func (l *LockSystem) Confirm(now time.Time, name0 string, name1 string, conditions ...webdav.Condition) (release func(), err error) {
	return l.Confirm(now, name0, name1, conditions...)
}

func (l *LockSystem) Create(now time.Time, details webdav.LockDetails) (token string, err error) {
	token = fmt.Sprintf("urn:%s", uuid.New().String())
	l.locks[token] = details
	return
}

func (l *LockSystem) Refresh(now time.Time, token string, duration time.Duration) (webdav.LockDetails, error) {
	return l.locks[token], nil
}

func (l *LockSystem) Unlock(now time.Time, token string) error {
	delete(l.locks, token)
	return nil
}
