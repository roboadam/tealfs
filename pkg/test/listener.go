package test

import "net"

type TestListener struct {
	listener net.Listener
}

func NewTestListener() *TestListener {
	listener, _ := net.Listen("tcp", "127.0.0.1:0")
	return &TestListener{listener}
}

func (testListener *TestListener) GetAddress() string {
	return testListener.listener.Addr().String()
}

func (testListener *TestListener) ReceivedConnection() bool {
	_, err := testListener.listener.Accept()
	return err == nil
}
