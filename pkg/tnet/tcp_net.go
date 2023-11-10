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
