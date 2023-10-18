package node_test

import (
	"bytes"
	"tealfs/pkg/cmds"
	"tealfs/pkg/node"
	"tealfs/pkg/test"
	"testing"
	"time"
)

func TestNodeCreation(t *testing.T) {
	userCmds := make(chan cmds.User)
	tNet := test.MockNet{}
	localNode := node.New(userCmds, &tNet)

	if !nodeIdIsValid(&localNode) {
		t.Error("Id is invalid")
	}
}

func TestConnectToRemoteNode(t *testing.T) {
	userCmds := make(chan cmds.User)
	tNet := test.MockNet{Dialed: false, AcceptsConnections: false}
	n := node.New(userCmds, &tNet)
	n.Start()

	userCmds <- cmds.User{CmdType: cmds.ConnectTo, Argument: "someAddress"}
	time.Sleep(time.Second * 5)

	if !tNet.IsDialed() {
		t.Error("Node did not connect")
	}

	if !bytes.Equal(tNet.Conn.BytesWritten, validHello(n.Id)) {
		t.Error("Node did not send valid hello")
	}
}

func TestIncomingConnection(t *testing.T) {
	userCmds := make(chan cmds.User)
	mockNet := test.MockNet{Dialed: false, AcceptsConnections: false}
	n := node.New(userCmds, &mockNet)
	n.Start()

	address := n.GetAddress()

	mockNet.
}

func validHello(nodeId node.Id) []byte {
	serializedHello := []byte{byte(int8(1))}
	serializedNodeId := []byte(nodeId.String())
	return append(serializedHello, serializedNodeId...)
}

func nodeIdIsValid(node *node.Node) bool {
	return len(node.Id.String()) > 0
}
