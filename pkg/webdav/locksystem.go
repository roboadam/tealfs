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
	"time"

	"golang.org/x/net/webdav"
)

type LockSystem struct{}

func (l *LockSystem) Confirm(now time.Time, name0 string, name1 string, conditions ...webdav.Condition) (release func(), err error) {
	panic("not implemented") // TODO: Implement
}

func (l *LockSystem) Create(now time.Time, details webdav.LockDetails) (token string, err error) {
	panic("not implemented") // TODO: Implement
}

func (l *LockSystem) Refresh(now time.Time, token string, duration time.Duration) (webdav.LockDetails, error) {
	panic("not implemented") // TODO: Implement
}

func (l *LockSystem) Unlock(now time.Time, token string) error {
	panic("not implemented") // TODO: Implement
}
