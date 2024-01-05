package mgr_test

import (
	"bytes"
	"encoding/binary"
	"errors"
	"os"
	"strconv"
	"tealfs/pkg/mgr"
	"tealfs/pkg/model/events"
	"tealfs/pkg/model/node"
	"tealfs/pkg/test"
	"tealfs/pkg/util"
	"testing"
	"time"
)

func TestManagerCreation(t *testing.T) {
	userCmds := make(chan events.Event)
	tNet := test.MockNet{}
	localNode := mgr.New(userCmds, &tNet)

	if !nodeIdIsValid(&localNode) {
		t.Error("Id is invalid")
	}
}

func TestConnectToRemoteNode(t *testing.T) {
	userCmds := make(chan events.Event)
	tNet := test.MockNet{Dialed: false, AcceptsConnections: false}
	n := mgr.New(userCmds, &tNet)
	n.Start()

	userCmds <- events.NewString(events.ConnectTo, "someAddress")

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
	userCmds := make(chan events.Event)
	mockNet := test.MockNet{Dialed: false, AcceptsConnections: true}
	n := mgr.New(userCmds, &mockNet)
	n.Start()

	remoteNodeId := node.NewNodeId()
	mockNet.Conn.SendMockBytes(validHello(remoteNodeId))

	expected := validHello(n.GetId())
	time.Sleep(time.Millisecond * 100)

	result := readPayloadBytesWritten(&mockNet)
	if !bytes.Equal(result, expected) {
		t.Error("You didn't hello back!")
	}
}

func TestSendNodeSyncAfterReceiveHello(t *testing.T) {
	userCmds := make(chan events.Event)
	tNet := test.MockNet{Dialed: false, AcceptsConnections: false}
	n := mgr.New(userCmds, &tNet)
	remoteNodeId := node.NewNodeId()
	remoteNodeAddress := "remoteAddress"
	n.Start()

	userCmds <- events.NewString(events.ConnectTo, remoteNodeAddress)

	if !tNet.IsDialed() {
		t.Error("Node did not connect")
	}

	time.Sleep(time.Millisecond * 50)

	tNet.Conn.BytesWritten = make([]byte, 0)
	tNet.Conn.SendMockBytes(validHello(remoteNodeId))

	expected := CommandAndNodes{Command: 2, Nodes: util.NewSet[NodeInfo]()}
	expected.Nodes.Add(NodeInfo{NodeId: remoteNodeId.String(), Address: remoteNodeAddress})
	expected.Nodes.Add(NodeInfo{NodeId: n.GetId().String(), Address: tNet.GetBinding()})

	time.Sleep(time.Millisecond * 20)
	payload := readPayloadBytesWritten(&tNet)
	commandAndNodes, err := commandAndNodesFrom(payload)
	if err != nil {
		t.Error(err.Error())
		return
	}

	if commandAndNodes.Command != expected.Command {
		t.Error("Invalid command " + strconv.Itoa(int(commandAndNodes.Command)))
	}

	if !commandAndNodes.Nodes.Equal(&expected.Nodes) {
		t.Error("Node set is not correct")
	}
}

func TestSave(t *testing.T) {
	userCmds := make(chan events.Event)
	tNet := test.MockNet{Dialed: false, AcceptsConnections: false}
	n := mgr.New(userCmds, &tNet)
	n.Start()

	tempDir, _ := os.MkdirTemp("", "*-test-save-mgr")
	defer removeAll(tempDir, t)

	userCmds <- events.NewString(events.AddStorage, tempDir)
	time.Sleep(100 * time.Millisecond)
	userCmds <- events.NewString(events.AddData, tempDir)
}

func validHello(nodeId node.Id) []byte {

	serializedHello := int8Serialized(1)
	serializedNodeId := []byte(nodeId.String())
	serializedNodeIdLen := intSerialized(len(serializedNodeId))

	payload := append(append(serializedHello, serializedNodeIdLen...), serializedNodeId...)
	serializedPayload := intSerialized(len(payload))

	return append(serializedPayload, payload...)
}

func readPayloadBytesWritten(tnet *test.MockNet) []byte {
	if len(tnet.Conn.BytesWritten) < 4 {
		return make([]byte, 0)
	}
	payloadLen := binary.BigEndian.Uint32(tnet.Conn.BytesWritten)
	returnVal := tnet.Conn.BytesWritten[:4+payloadLen]
	tnet.Conn.BytesWritten = tnet.Conn.BytesWritten[4+payloadLen:]
	return returnVal
}

type NodeInfo struct {
	NodeId  string
	Address string
}

type CommandAndNodes struct {
	Command int8
	Nodes   util.Set[NodeInfo]
}

func commandAndNodesFrom(data []byte) (*CommandAndNodes, error) {
	length := binary.BigEndian.Uint32(data)
	if int(length) != len(data)-4 {
		return nil, errors.New("invalid length")
	}
	command := int8(data[4])
	if command != 2 {
		return nil, errors.New("not a SyncNodes")
	}
	nodes := util.NewSet[NodeInfo]()

	start := 5

	for start < len(data) {
		idLen := int(binary.BigEndian.Uint32(data[start:]))
		start += 4
		id := string(data[start : start+idLen])
		start += idLen
		addressLen := int(binary.BigEndian.Uint32(data[start:]))
		start += 4
		address := string(data[start : start+addressLen])
		start += addressLen
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

func nodeIdIsValid(mgr *mgr.Mgr) bool {
	return len(mgr.GetId().String()) > 0
}

func removeAll(dir string, t *testing.T) {
	err := os.RemoveAll(dir)
	if err != nil {
		t.Errorf("Error [%v] deleting temp dir [%v]", err, dir)
	}
}
