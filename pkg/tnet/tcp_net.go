package tnet

import (
	"net"
	"time"
)

type TcpNet struct {
	address  string
	listener net.Listener
}

func NewTcpNet(address string) *TcpNet {
	return &TcpNet{address: address}
}

func (t *TcpNet) Dial() net.Conn {
	conn, err := net.Dial("tcp", t.address)
	for err != nil {
		time.Sleep(time.Second * 2)
		conn, err = net.Dial("tcp", t.address)
	}
	return conn
}

func (t *TcpNet) listen() {
	listener, err := net.Listen("tcp", t.address)
	for err != nil {
		time.Sleep(time.Second * 2)
		listener, err = net.Listen("tcp", t.address)
	}
	t.listener = listener
}

func (t *TcpNet) GetAddress() string {
	return t.address
}

func (t *TcpNet) SetAddress(address string) {
	t.address = address
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
