package tnet

import "net"

type TNet interface {
	Dial(address string) net.Conn
	BindTo(address string)
	Close()
	Accept() net.Conn
}
