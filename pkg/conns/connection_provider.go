// Copyright (C) 2025 Adam Hess
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

package conns

import (
	"bytes"
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
	Conn     MockConn

	DialedAddress chan string
}

func NewMockConnectionProvider() MockConnectionProvider {
	return MockConnectionProvider{
		Listener: NewMockListener(),
		Conn: MockConn{
			dataToRead:  make(chan ClosableBuffer, 1),
			dataWritten: ClosableBuffer{},
		},
		DialedAddress: make(chan string, 1),
	}
}

func (m *MockConnectionProvider) GetConnection(address string) (net.Conn, error) {
	m.DialedAddress <- address
	return &m.Conn, nil
}

func (m *MockConnectionProvider) GetListener(address string) (net.Listener, error) {
	return &m.Listener, nil
}

type MockConn struct {
	dataToRead  chan ClosableBuffer
	dataWritten ClosableBuffer
	buffPtr     *ClosableBuffer
	ReadError   error
	WriteError  error
}
type MockAddr struct{}

type ClosableBuffer bytes.Buffer

func (b *ClosableBuffer) Close() error {
	return nil
}

func (b *ClosableBuffer) Write(p []byte) (n int, err error) {
	return (*bytes.Buffer)(b).Write(p)
}

func (b *ClosableBuffer) Read(p []byte) (n int, err error) {
	return (*bytes.Buffer)(b).Read(p)
}

func (m *MockConn) Read(b []byte) (n int, err error) {
	if m.ReadError != nil {
		return 0, m.ReadError
	}
	if m.buffPtr == nil {
		buf := <-m.dataToRead
		m.buffPtr = &buf
	}
	return m.buffPtr.Read(b)
}

func (m *MockConn) Write(b []byte) (n int, err error) {
	if m.WriteError != nil {
		return 0, m.WriteError
	}
	return m.dataWritten.Write(b)
}

func (m *MockConn) Close() error {
	return nil
}

func (m *MockConn) LocalAddr() net.Addr {
	return MockAddr{}
}

func (m *MockConn) RemoteAddr() net.Addr {
	return MockAddr{}
}

func (m *MockConn) SetDeadline(t time.Time) error {
	panic("not implemented6") // TODO: Implement
}

func (m *MockConn) SetReadDeadline(t time.Time) error {
	panic("not implemented7") // TODO: Implement
}

func (m *MockConn) SetWriteDeadline(t time.Time) error {
	panic("not implemented8") // TODO: Implement
}

func (m MockAddr) Network() string {
	return "tcp"
}

func (m MockAddr) String() string {
	return "mockAddress:123"
}

type MockListener struct {
	accept chan bool
}

func NewMockListener() MockListener {
	return MockListener{
		accept: make(chan bool),
	}
}

func (m *MockListener) Accept() (net.Conn, error) {
	<-m.accept
	return &MockConn{
		dataToRead:  make(chan ClosableBuffer, 100),
		dataWritten: ClosableBuffer{},
	}, nil
}

func (m *MockListener) Close() error {
	return nil
}

func (m *MockListener) Addr() net.Addr {
	return MockAddr{}
}
