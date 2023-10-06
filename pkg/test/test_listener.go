package test

import "net"

type TestListener struct {
	listener     net.Listener
	savedAddress string
}

func NewTestListener() *TestListener {
	listener, _ := net.Listen("tcp", "127.0.0.1:0")
	savedAddress := listener.Addr().String()
	return &TestListener{listener, savedAddress}
}

func (testListener *TestListener) GetAddress() string {
	return testListener.listener.Addr().String()
}

func (testListener *TestListener) ReceivedConnection() bool {
	_, err := testListener.listener.Accept()
	return err == nil
}

func (listener *TestListener) CloseAndReopen() {
	listener.listener.Close()
	listener.listener, _ = net.Listen("tcp", listener.savedAddress)
}
