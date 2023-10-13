package tnet

import "net"

type TcpNet struct{}

func (tcpNet *TcpNet) Dial(network, address string) (net.Conn, error) {
	return net.Dial(network, address)
}
