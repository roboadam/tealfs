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

package tnet

import (
	"net"
	"time"
)

type TcpNet struct {
	listener net.Listener
	binding  string
}

func NewTcpNet(binding string) *TcpNet {
	result := TcpNet{binding: binding}
	result.listen()
	return &result
}

func (t *TcpNet) Dial(address string) net.Conn {
	conn, err := net.Dial("tcp", address)
	for err != nil {
		time.Sleep(time.Second * 2)
		conn, err = net.Dial("tcp", address)
	}
	return conn
}

func (t *TcpNet) listen() {
	listener, err := net.Listen("tcp", t.binding)
	for err != nil {
		time.Sleep(time.Second * 2)
		listener, err = net.Listen("tcp", t.binding)
	}
	t.binding = listener.Addr().String()
	t.listener = listener
}

func (t *TcpNet) Close() {
	if t.listener != nil {
		t.Close()
	}
}

func (t *TcpNet) Accept() net.Conn {
	if t.listener == nil {
		t.listen()
	}

	conn, err := t.listener.Accept()
	for err != nil {
		time.Sleep(time.Second * 2)
		conn, err = t.listener.Accept()
	}

	return conn
}

func (t *TcpNet) GetBinding() string {
	return t.binding
}
