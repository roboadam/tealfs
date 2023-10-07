package node

import (
	"errors"
	"net"
)

type RemoteNode struct {
	NodeId  Id
	Address string
	Conn    net.Conn
}

func (node *RemoteNode) Connect() {
	if node.Conn == nil {
		node.connectUnconnectedNode()
	}
}

func (node *RemoteNode) Disconnect() {
	if node.Conn != nil {
		node.Conn.Close()
	}
}

func (node *RemoteNode) connectUnconnectedNode() error {
	tcpConn, err := net.Dial("tcp", node.Address)
	if err == nil {
		node.Conn = tcpConn
	} else {
		return errors.New("can't connect")
	}
	return nil
}
