package test

import "net"

type Listener struct {
	Accepted bool
	Closed   bool
}

type Addr struct {
}

func (a Addr) Network() string {
	return "tcp"
}

func (a Addr) String() string {
	return "127.0.0.1"
}

func (listener *Listener) Accept() (net.Conn, error) {
	listener.Accepted = true
	return TestConn{}, nil
}

func (listener *Listener) Close() error {
	listener.Closed = true
	return nil
}

func (listener *Listener) Addr() net.Addr {
	return Addr{}
}
