package node

import (
	"tealfs/pkg/tnet"
)

type RemoteNode struct {
	NodeId  Id
	Address string
	tNet    tnet.TNet
}

func NewRemoteNode(nodeId Id, address string, tNet tnet.TNet) *RemoteNode {
	return &RemoteNode{NodeId: nodeId, tNet: tNet}
}

func (node *RemoteNode) Connect() {
	node.tNet.Dial(node.Address)
}

func (node *RemoteNode) Disconnect() {
	node.tNet.Close()
}
