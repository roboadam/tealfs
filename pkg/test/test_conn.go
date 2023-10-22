package test

import (
	"net"
	"time"
)

type Conn struct {
	BytesWritten []byte
	BytesToRead  []byte
}

func (m *Conn) Read(b []byte) (n int, err error) {
	copy(b, m.BytesToRead)
	m.BytesToRead = m.BytesToRead[len(b):]
	return len(b), nil
}

func (m *Conn) Write(b []byte) (n int, err error) {
	m.BytesWritten = append(m.BytesWritten, b...)
	return len(b), nil
}

func (m *Conn) Close() error {
	return nil
}

func (m *Conn) LocalAddr() net.Addr {
	return &net.IPAddr{IP: net.IPv4(127, 0, 0, 1)}
}

func (m *Conn) RemoteAddr() net.Addr {
	return &net.IPAddr{IP: net.IPv4(192, 168, 0, 1)}
}

func (m *Conn) SetDeadline(time.Time) error {
	return nil
}

func (m *Conn) SetReadDeadline(time.Time) error {
	return nil
}

func (m *Conn) SetWriteDeadline(time.Time) error {
	return nil
}

func (m *Conn) SendMockBytes(hello []byte) {
	m.BytesToRead = append(m.BytesToRead, hello...)
}
