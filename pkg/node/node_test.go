package node_test

import (
	"tealfs/pkg/cmds"
	"tealfs/pkg/node"
	"tealfs/pkg/test"
	"testing"
)

func TestNodeCreation(t *testing.T) {
	userCmds := make(chan cmds.User)
	node := listeningNode(userCmds)
	defer node.Close()

	if !nodeIdisValid(node) {
		t.Error("NodeId is invalid")
	}

	if !nodeAddressIsValid(node) {
		t.Error("Node address is invalid")
	}
}

func TestConnectToRemoteNode(t *testing.T) {
	userCmds := make(chan cmds.User)
	node := listeningNode(userCmds)
	defer node.Close()
	testListener := test.NewTestListener()

	userCmds <- cmds.User{CmdType: cmds.ConnectTo, Argument: testListener.GetAddress()}

	if !testListener.ReceivedConnection() {
		t.Error("Node did not connect")
	}
}

func listeningNode(userCmds chan cmds.User) *node.Node {
	node := node.NewNode(userCmds)
	node.SetHostToBind("127.0.0.1")
	node.Listen()
	return &node
}

func nodeIdisValid(node *node.Node) bool {
	return len(node.Id.String()) > 0
}

func nodeAddressIsValid(node *node.Node) bool {
	return len(node.GetAddress().String()) > 0
}
