package tnet

import "net"

type TNet interface {
	Dial() net.Conn
	GetAddress() string
	SetAddress(string)
	Close()
	Accept() net.Conn
}
