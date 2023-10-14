package node

import (
	"net"
	"tealfs/pkg/tnet"
	"time"
)

type RemoteNode struct {
	NodeId  Id
	Address string
	TNet    tnet.TNet
}

func (node *RemoteNode) Connect() {
	if node.Conn == nil {
		node.connectUntilSuccess()
	}
}

func (node *RemoteNode) Disconnect() {
	if node.Conn != nil {
		node.Conn.Close()
	}
}

func (node *RemoteNode) connectUntilSuccess() {
	var err error
	node.Conn, err = net.Dial("tcp", node.Address)
	for err == nil {
		time.Sleep(time.Second * 2)
		node.Conn, err = net.Dial("tcp", node.Address)
	}
}
