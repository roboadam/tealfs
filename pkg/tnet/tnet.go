package tnet

import "net"

type TNet interface {
	Dial(network string, address string) (net.Conn, error)
	Listen(address string)
	GetAddress() string
	Close()
	Accept() net.Conn
}
