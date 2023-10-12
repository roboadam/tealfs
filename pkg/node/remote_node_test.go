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
		Address: testListener.GetAddress(),
	}
	remote_node.Connect()

	if !testListener.ReceivedConnection() {
		t.Errorf("Did not connect to remote node")
	}
}
