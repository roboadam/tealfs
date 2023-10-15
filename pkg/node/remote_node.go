package node

import (
	"tealfs/pkg/tnet"
)

type RemoteNode struct {
	NodeId Id
	tNet   tnet.TNet
}

func NewRemoteNode(nodeId Id, tNet tnet.TNet) *RemoteNode {
	return &RemoteNode{NodeId: nodeId, tNet: tNet}
}

func (node *RemoteNode) Connect() {
	node.tNet.Dial()
}

func (node *RemoteNode) Disconnect() {
	node.tNet.Close()
}
