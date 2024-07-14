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
	GetListener(address string) (net.Listener, error)
}

type TcpConnectionProvider struct{}

func (r *TcpConnectionProvider) GetConnection(address string) (net.Conn, error) {
	return net.Dial("tcp", address)
}

func (r *TcpConnectionProvider) GetListener(address string) (net.Listener, error) {
	return net.Listen("tcp", address)
}

type MockConnectionProvider struct {
	Listener MockListener
}

func (m MockConnectionProvider) GetConnection(address string) (net.Conn, error) {
	return MockConn{}, nil
}

func (m MockConnectionProvider) GetListener(address string) (net.Listener, error) {
	return MockListener{}, nil
}

type MockConn struct{}
type MockAddr struct{}

func (m MockConn) Read(b []byte) (n int, err error) {
	panic("not implemented1") // TODO: Implement
}

func (m MockConn) Write(b []byte) (n int, err error) {
	panic("not implemented2") // TODO: Implement
}

func (m MockConn) Close() error {
	panic("not implemented3") // TODO: Implement
}

func (m MockConn) LocalAddr() net.Addr {
	panic("not implemented4") // TODO: Implement
}

func (m MockConn) RemoteAddr() net.Addr {
	panic("not implemented5") // TODO: Implement
}

func (m MockConn) SetDeadline(t time.Time) error {
	panic("not implemented6") // TODO: Implement
}

func (m MockConn) SetReadDeadline(t time.Time) error {
	panic("not implemented7") // TODO: Implement
}

func (m MockConn) SetWriteDeadline(t time.Time) error {
	panic("not implemented8") // TODO: Implement
}

func (m MockAddr) Network() string {
	return "tcp"
}

func (m MockAddr) String() string {
	return "mockaddress:123"
}

type MockListener struct {
	accept chan bool
}

func (m MockListener) Accept() (net.Conn, error) {
	<-m.accept
	return MockConn{}, nil
}

func (m MockListener) Close() error {
	panic("not implemented9") // TODO: Implement
}

func (m MockListener) Addr() net.Addr {
	return MockAddr{}
}
