package test

import (
	"net"
	"tealfs/pkg/tnet"
	"time"
)

type TestNet struct {
	Accepted           bool
	Dialed             bool
	AcceptsConnections bool
}

func (t TestNet) Dial(address string) net.Conn {
	t.Dialed = true
	return TestConn{}
}

func (t TestNet) IsDialed() bool {
	for !t.Dialed {
		println("Sleeping one second")
		time.Sleep(time.Second)
	}
	return true
}

func (t TestNet) BindTo(address string) {
}

func (t TestNet) Close() {
}

func (t TestNet) Accept() net.Conn {
	t.Accepted = true
	if !t.AcceptsConnections {
		for {
			time.Sleep(time.Minute)
		}
	}
	return TestConn{}
}

func iFace(t tnet.TNet) {

}

func testIface() {
	t := TestNet{}
	iFace(t)
}
