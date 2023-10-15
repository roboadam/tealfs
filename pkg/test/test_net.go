package test

import "net"

type TestNet struct {
	Accepted bool
}

func (t *TestNet) Dial() net.Conn {
	return TestConn{}
}

func (t *TestNet) GetAddress() string {
	return "someaddress"
}

func (t *TestNet) SetAddress(s string) {
}

func (t *TestNet) Close() {
}

func (t *TestNet) Accept() net.Conn {
	t.Accepted = true
	return TestConn{}
}
