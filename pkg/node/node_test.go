package node_test

import (
	"tealfs/pkg/cmds"
	"tealfs/pkg/node"
	"tealfs/pkg/test"
	"testing"
)

func TestNodeCreation(t *testing.T) {
	userCmds := make(chan cmds.User)
	tNet := test.TestNet{}
	localNode := node.New(userCmds, &tNet)

	if !nodeIdIsValid(&localNode) {
		t.Error("Id is invalid")
	}

	if !nodeAddressIsValid(&localNode) {
		t.Error("Node address is invalid")
	}
}

func TestConnectToRemoteNode(t *testing.T) {
	userCmds := make(chan cmds.User)
	tNet := test.TestNet{Dialed: false, AcceptsConnections: false}
	_ = node.New(userCmds, &tNet)

	userCmds <- cmds.User{CmdType: cmds.ConnectTo, Argument: "someAddress"}

	if !tNet.IsDialed() {
		t.Error("Node did not connect")
	}
}

func nodeIdIsValid(node *node.Node) bool {
	return len(node.Id.String()) > 0
}

func nodeAddressIsValid(node *node.Node) bool {
	return len(node.GetAddress()) > 0
}
