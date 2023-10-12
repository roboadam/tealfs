package test

import (
	"net"
	"time"
)

type TestConn struct {
}

func (m TestConn) Read(b []byte) (n int, err error) {
	copy(b, []byte("Hello, World!"))
	return len("Hello, World!"), nil
}

func (m TestConn) Write(b []byte) (n int, err error) {
	return len(b), nil
}

func (m TestConn) Close() error {
	return nil
}

func (m TestConn) LocalAddr() net.Addr {
	return &net.IPAddr{IP: net.IPv4(127, 0, 0, 1)}
}

func (m TestConn) RemoteAddr() net.Addr {
	return &net.IPAddr{IP: net.IPv4(192, 168, 0, 1)}
}

func (m TestConn) SetDeadline(t time.Time) error {
	return nil
}

func (m TestConn) SetReadDeadline(t time.Time) error {
	return nil
}

func (m TestConn) SetWriteDeadline(t time.Time) error {
	return nil
}
