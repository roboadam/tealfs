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
	defer localNode.Close()

	if !nodeIdIsValid(localNode) {
		t.Error("Id is invalid")
	}

	if !nodeAddressIsValid(localNode) {
		t.Error("Node address is invalid")
	}
}

func TestConnectToRemoteNode(t *testing.T) {
	userCmds := make(chan cmds.User)
	tNet := test.TestNet{}
	localNode := node.New(userCmds, &tNet)
	defer localNode.Close()

	userCmds <- cmds.User{CmdType: cmds.ConnectTo, Argument: "someAddress"}

	if !tNet.Accepted {
		t.Error("Node did not connect")
	}
}

func nodeIdIsValid(node *node.Node) bool {
	return len(node.Id.String()) > 0
}

func nodeAddressIsValid(node *node.Node) bool {
	return len(node.GetAddress()) > 0
}
