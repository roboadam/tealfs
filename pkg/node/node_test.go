package node_test

import (
	"tealfs/pkg/cmds"
	"tealfs/pkg/node"
	"testing"
)

func TestNodeCreation(t *testing.T) {
	userCmds := make(chan cmds.User)
	node := node.NewNode(userCmds)
	node.SetHostToBind("127.0.0.1")
	node.Listen()
	defer node.Close()

	if !nodeIdisValid(&node) {
		t.Error("NodeId is invalid")
	}

	if !nodeAddressIsValid(&node) {
		t.Error("Node address is invalid")
	}
}

func nodeIdisValid(node *node.Node) bool {
	return len(node.Id.String()) > 0
}

func nodeAddressIsValid(node *node.Node) bool {
	return len(node.GetAddress().String()) > 0
}
