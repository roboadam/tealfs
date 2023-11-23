package mgr_test

import (
	"bytes"
	"encoding/binary"
	"strconv"
	"tealfs/pkg/cmds"
	"tealfs/pkg/mgr"
	"tealfs/pkg/node"
	"tealfs/pkg/test"
	"tealfs/pkg/util"
	"testing"
	"time"
)

func TestManagerCreation(t *testing.T) {
	userCmds := make(chan cmds.User)
	tNet := test.MockNet{}
	localNode := mgr.New(userCmds, &tNet)

	if !nodeIdIsValid(&localNode) {
		t.Error("Id is invalid")
	}
}

func TestConnectToRemoteNode(t *testing.T) {
	userCmds := make(chan cmds.User)
	tNet := test.MockNet{Dialed: false, AcceptsConnections: false}
	n := mgr.New(userCmds, &tNet)
	n.Start()

	userCmds <- cmds.User{CmdType: cmds.ConnectTo, Argument: "someAddress"}

	if !tNet.IsDialed() {
		t.Error("Node did not connect")
	}

	expected := validHello(n.GetId())

	time.Sleep(time.Millisecond * 100)
	if !bytes.Equal(tNet.Conn.BytesWritten, expected) {
		t.Errorf("Node did not send valid hello, %d %d", len(tNet.Conn.BytesWritten), len(expected))
	}
}

func TestIncomingConnection(t *testing.T) {
	userCmds := make(chan cmds.User)
	mockNet := test.MockNet{Dialed: false, AcceptsConnections: true}
	n := mgr.New(userCmds, &mockNet)
	n.Start()

	remoteNodeId := node.NewNodeId()
	mockNet.Conn.SendMockBytes(validHello(remoteNodeId))

	expected := validHello(n.GetId())
	time.Sleep(time.Millisecond * 100)
	if !bytes.Equal(mockNet.Conn.BytesWritten, expected) {
		t.Error("You didn't hello back!")
	}

}

func TestSendNodeSyncAfterReceiveHello(t *testing.T) {
	userCmds := make(chan cmds.User)
	mockNet := test.MockNet{Dialed: false, AcceptsConnections: true}
	n := mgr.New(userCmds, &mockNet)
	n.Start()

	remoteNodeId := node.NewNodeId()
	mockNet.Conn.SendMockBytes(validHello(remoteNodeId))

	expected := CommandAndNodes{Command: 2, Nodes: util.NewSet[NodeInfo]()}
	expected.Nodes.Add(NodeInfo{NodeId: remoteNodeId.String(), Address: "something"})
	expected.Nodes.Add(NodeInfo{NodeId: n.GetId().String(), Address: "something else"})

	commandAndNodes, err := CommandAndNodesFrom(mockNet.Conn.BytesWritten)
	if err != nil {
		t.Error(err.Error())
	}

	if commandAndNodes.Command != 2 {
		t.Error("Invalid command " + strconv.Itoa(int(commandAndNodes.Command)))
	}

	if !commandAndNodes.Nodes.Equal(&expected.Nodes) {
		t.Error("Node set is not correct")
	}
}

func TestSendNodeSyncAfterReceiveNodeSync(t *testing.T) {

}

func TestReceiveNodeSyncAddsMissingNodes(t *testing.T) {

}

func validHello(nodeId node.Id) []byte {

	serializedHello := int8Serialized(1)
	serializedNodeId := []byte(nodeId.String())
	seralizedNodeIdLen := intSerialized(len(serializedNodeId))

	payload := append(append(serializedHello, seralizedNodeIdLen...), serializedNodeId...)
	seralizedPayoadLen := intSerialized(len(payload))

	return append(seralizedPayoadLen, payload...)
}

type NodeInfo struct {
	NodeId  string
	Address string
}

type CommandAndNodes struct {
	Command int8
	Nodes   util.Set[NodeInfo]
}

func CommandAndNodesFrom(data []byte) (*CommandAndNodes, error) {
	command := int8(data[0])
	nodes := util.NewSet[NodeInfo]()

	start := 1

	for start < len(data) {
		idLen := int(binary.BigEndian.Uint32(data[start:]))
		start += 4
		id := string(data[start : start+idLen])
		addressLen := int(binary.BigEndian.Uint32(data[start:]))
		address := string(data[start : start+addressLen])
		nodes.Add(NodeInfo{NodeId: id, Address: address})
	}

	return &CommandAndNodes{Command: command, Nodes: nodes}, nil
}

func intSerialized(number int) []byte {
	serializedInt := make([]byte, 4)
	binary.BigEndian.PutUint32(serializedInt, uint32(number))
	return serializedInt
}

func int8Serialized(number int8) []byte {
	return []byte{byte(number)}
}

func nodeIdIsValid(mgr *mgr.Manager) bool {
	return len(mgr.GetId().String()) > 0
}
