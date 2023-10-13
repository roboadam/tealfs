package node_test

import (
	"tealfs/pkg/node"
	"tealfs/pkg/test"
	"testing"
)

func TestConnect(t *testing.T) {
	testListener := test.Listener{Accepted: false, Closed: false}
	remote_node := node.RemoteNode{
		NodeId:  node.NewNodeId(),
		Address: "someaddress",
	}
	remote_node.Connect()

	if !testListener.Accepted {
		t.Errorf("Did not connect to remote node")
	}
}
