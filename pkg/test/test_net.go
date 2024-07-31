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

package test

import (
	"net"
	"time"
)

type MockNet struct {
	Accepted           bool
	Dialed             bool
	AcceptsConnections bool
	Conn               Conn
}

func (t *MockNet) Dial(string) net.Conn {
	t.Dialed = true
	t.Conn = Conn{BytesWritten: make([]byte, 0)}
	return &t.Conn
}

func (t *MockNet) Close() {
}

func (t *MockNet) Accept() net.Conn {
	t.Accepted = true
	if !t.AcceptsConnections {
		for {
			time.Sleep(time.Minute)
		}
	} else {
		t.AcceptsConnections = false
	}
	return &t.Conn
}

func (t *MockNet) GetBinding() string {
	return "mockbinding:123"
}

func (t *MockNet) IsDialed() bool {
	for !t.Dialed {
		time.Sleep(time.Millisecond * 10)
	}
	return true
}
