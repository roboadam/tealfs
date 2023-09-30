package node_test

import (
	"tealfs/pkg/cmds"
	"tealfs/pkg/node"
	"testing"
)

func TestConstructor(t *testing.T) {
	userCmds := make(chan cmds.User)
	node := node.NewNode(userCmds)
	node.Start()

	if !nodeIdExists(&node) {
		t.Error("NodeId does not exist")
	}

	if !nodeAddressIsGood(&node)
}

func nodeIdExists(node *node.Node) bool {
	return len(node.Id.String()) > 0
}

func nodeAddressIsGood(node *node.Node) bool {
	return len(node.GetAddress().String()) > 0
}
