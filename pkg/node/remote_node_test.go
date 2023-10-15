package node_test

import (
	"tealfs/pkg/node"
	"tealfs/pkg/test"
	"testing"
)

func TestConnect(t *testing.T) {
	testNet := test.TestNet{Dialed: false}
	remoteNode := node.NewRemoteNode(node.NewNodeId(), &testNet)
	remoteNode.Connect()

	if !testNet.Dialed {
		t.Errorf("Did not dial")
	}
}
