package test

import "net"

type TestNet struct{}

func (testNet *TestNet) Dial(network, address string) (net.Conn, error) {
	return TestConn{}, nil
}
