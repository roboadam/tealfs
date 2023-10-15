package tnet

import "net"

type TNet interface {
	Dial(address string) net.Conn
	BindTo(string)
	Close()
	Accept() net.Conn
}
