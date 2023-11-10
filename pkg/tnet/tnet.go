package tnet

import "net"

type TNet interface {
	Dial(address string) net.Conn
	Close()
	Accept() net.Conn
	GetBinding() string
}
