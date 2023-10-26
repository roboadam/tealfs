package node_test

import (
	"bytes"
	"encoding/binary"
	"tealfs/pkg/cmds"
	"tealfs/pkg/node"
	"tealfs/pkg/test"
	"testing"
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

	if !tNet.IsDialed() {
		t.Error("Node did not connect")
	}

	if !bytes.Equal(tNet.Conn.BytesWritten, validHello(n.Id)) {
		t.Error("Node did not send valid hello")
	}
}

func TestIncomingConnection(t *testing.T) {
	userCmds := make(chan cmds.User)
	mockNet := test.MockNet{Dialed: false, AcceptsConnections: true}
	n := node.New(userCmds, &mockNet)
	n.Start()

	remoteNodeId := node.NewNodeId()
	mockNet.Conn.SendMockBytes(validHello(remoteNodeId))

	remoteNode, err := n.GetRemoteNode(remoteNodeId)
	if err != nil || remoteNode == nil || remoteNode.Id != remoteNodeId {
		t.Error("Did not add node " + remoteNodeId.String() + " to cluster")
	}
}

func validHello(nodeId node.Id) []byte {
	serializedHello := []byte{byte(int8(1))}
	serializedNodeIdLen := make([]byte, 4)
	binary.BigEndian.PutUint32(serializedNodeIdLen, uint32(len(nodeId.String())))
	serializedNodeId := []byte(nodeId.String())
	return append(append(serializedHello, serializedNodeIdLen...), serializedNodeId...)
}

type NodeInfo struct {
	NodeId  string
	Address string
}

func nodeSync(nodes NodeInfo[]) []byte {
	return make([]byte, 0)
}

func stringSerialized(value string) []byte {
	serializedNodeIdLen := make([]byte, 4)
	binary.BigEndian.PutUint32(serializedNodeIdLen, uint32(len(value)))
	serializedNodeId := []byte(value)
	return append(serializedNodeIdLen, serializedNodeId...)
}

func nodeIdIsValid(node *node.Node) bool {
	return len(node.Id.String()) > 0
}
