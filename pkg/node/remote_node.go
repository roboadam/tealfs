package node

import "net"

type RemoteNode struct {
	NodeId  NodeId
	Address string
	Conn    net.Conn
}

