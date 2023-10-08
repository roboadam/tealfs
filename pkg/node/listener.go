package node

import (
	"errors"
	"net"
)

type Listener struct {
	value net.Listener
}

func (listener *Listener) GetAddress() string {
	if listener.value == nil {
		return ""
	}
	return listener.value.Addr().String()
}

func (listener *Listener) Close() {
	if listener.value != nil {
		_ = listener.value.Close()
	}
}

func (listener *Listener) ListenOnFreePort(bind string) error {
	var err error
	listener.value, err = net.Listen("tcp", bind+":0")
	return err
}

func (listener *Listener) Accept() (net.Conn, error) {
	if listener.value != nil {
		return listener.value.Accept()
	}
	return nil, errors.New("listener is null")
}
