package tnet

import "net"

type TNet interface {
	Dial(network, address string) (net.Conn, error)
}
