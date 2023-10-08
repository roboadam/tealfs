package test

import "net"

type Listener struct {
	listener     net.Listener
	savedAddress string
}

func NewTestListener() *Listener {
	listener, _ := net.Listen("tcp", "127.0.0.1:0")
	savedAddress := listener.Addr().String()
	return &Listener{listener, savedAddress}
}

func (listener *Listener) GetAddress() string {
	return listener.listener.Addr().String()
}

func (listener *Listener) ReceivedConnection() bool {
	_, err := listener.listener.Accept()
	return err == nil
}
