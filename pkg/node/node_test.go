package node_test

import (
	"tealfs/pkg/cmds"
	"tealfs/pkg/node"
	"tealfs/pkg/test"
	"testing"
)

func TestNodeCreation(t *testing.T) {
	userCmds := make(chan cmds.User)
	localNode := listeningNode(userCmds)
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
	localNode := listeningNode(userCmds)
	defer localNode.Close()
	testListener := test.Listener{Accepted: false, Closed: false}

	userCmds <- cmds.User{CmdType: cmds.ConnectTo, Argument: "someAddress"}

	if !testListener.Accepted {
		t.Error("Node did not connect")
	}
}

func listeningNode(userCmds chan cmds.User) *node.Node {
	localNode := node.New(userCmds)
	localNode.HostToBind = "127.0.0.1"
	localNode.Listen()
	return &localNode
}

func nodeIdIsValid(node *node.Node) bool {
	return len(node.Id.String()) > 0
}

func nodeAddressIsValid(node *node.Node) bool {
	return len(node.GetAddress()) > 0
}
