package test

import (
	"net"
	"time"
)

type TestNet struct {
	Accepted           bool
	Dialed             bool
	AcceptsConnections bool
}

func (t *TestNet) Dial(_ string) net.Conn {
	t.Dialed = true
	return TestConn{}
}

func (t *TestNet) IsDialed() bool {
	for !t.Dialed {
		println("Sleeping one second")
		time.Sleep(time.Second)
	}
	return true
}

func (t *TestNet) BindTo(string) {
}

func (t *TestNet) Close() {
}

func (t *TestNet) Accept() net.Conn {
	t.Accepted = true
	if !t.AcceptsConnections {
		for {
			time.Sleep(time.Minute)
		}
	}
	return TestConn{}
}
