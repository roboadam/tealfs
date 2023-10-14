package tnet

import "net"

type TNet interface {
	Dial(network string, address string) (net.Conn, error)
	Listen(network string, address string) (net.Listener, error)
}
