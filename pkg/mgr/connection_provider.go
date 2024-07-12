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

package mgr

import (
	"net"
	"time"
)

type ConnectionProvider interface {
	GetConnection(address string) (net.Conn, error)
}

type TcpConnectionProvider struct{}

func (r *TcpConnectionProvider) GetConnection(address string) (net.Conn, error) {
	return net.Dial("tcp", address)
}

type MockConnectionProvider struct{}

func (m *MockConnectionProvider) GetConnection(address string) (net.Conn, error) {
	return nil, nil
}

type MockConn struct{}

func (m MockConn) Read(b []byte) (n int, err error) {
	panic("not implemented") // TODO: Implement
}

func (m MockConn) Write(b []byte) (n int, err error) {
	panic("not implemented") // TODO: Implement
}

func (m MockConn) Close() error {
	panic("not implemented") // TODO: Implement
}

func (m MockConn) LocalAddr() net.Addr {
	panic("not implemented") // TODO: Implement
}

func (m MockConn) RemoteAddr() net.Addr {
	panic("not implemented") // TODO: Implement
}

func (m MockConn) SetDeadline(t time.Time) error {
	panic("not implemented") // TODO: Implement
}

func (m MockConn) SetReadDeadline(t time.Time) error {
	panic("not implemented") // TODO: Implement
}

func (m MockConn) SetWriteDeadline(t time.Time) error {
	panic("not implemented") // TODO: Implement
}
