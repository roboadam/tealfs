package node

import (
	"net"
	"tealfs/pkg/raw_net"
	"tealfs/pkg/tnet"
)

type RemoteNode struct {
	Id      Id
	Address string
	tNet    tnet.TNet
	conn    net.Conn
}

func NewRemoteNode(nodeId Id, address string, tNet tnet.TNet) *RemoteNode {
	return &RemoteNode{Id: nodeId, Address: address, tNet: tNet}
}

func (r *RemoteNode) Connect() {
	r.conn = r.tNet.Dial(r.Address)
	_ = raw_net.Int8To(r.conn, 1)
	_ = raw_net.UInt32To(r.conn, uint32(len(r.Id.String())))
	_ = raw_net.StringTo(r.conn, r.Id.String())
}

func (r *RemoteNode) Disconnect() {
	r.tNet.Close()
}
