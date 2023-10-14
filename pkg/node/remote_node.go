package node

import (
	"tealfs/pkg/tnet"
)

type RemoteNode struct {
	nodeId Id
	tNet   tnet.TNet
}

func (node *RemoteNode) Connect() {
	node.tNet.Dial("tcp", node.tNet.GetAddress())
}

func (node *RemoteNode) Disconnect() {
	node.tNet.Close()
}
