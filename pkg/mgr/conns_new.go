package mgr

import "net"

type ConnsNew struct {
	netConns []net.Conn
}

func (c *ConnsNew) ConnectTo(address string) error {
	netConn, err := net.Dial("tcp", address)
	if err != nil {
		return err
	}
	c.netConns = append(c.netConns, netConn)
	return nil
}
