package test

import (
	"net"
	"time"
)

type MockNet struct {
	Accepted           bool
	Dialed             bool
	AcceptsConnections bool
}

func (t *MockNet) Dial(string) net.Conn {
	t.Dialed = true
	return Conn{}
}

func (t *MockNet) BindTo(string) {
}

func (t *MockNet) Close() {
}

func (t *MockNet) Accept() net.Conn {
	t.Accepted = true
	if !t.AcceptsConnections {
		for {
			time.Sleep(time.Minute)
		}
	}
	return Conn{}
}

func (t *MockNet) IsDialed() bool {
	for !t.Dialed {
		println("Sleeping one second")
		time.Sleep(time.Second)
	}
	return true
}
