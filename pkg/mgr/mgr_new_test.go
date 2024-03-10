package mgr

import (
	"fmt"
	"net"
	"os"
	"tealfs/pkg/store"
	"testing"
)

func TestConnectToRemoteNodeNew(t *testing.T) {
	dir := tmpDir()
	defer cleanDir(dir, t)
	m := NewNew()
	m.Start()

	listener, _ := net.Listen("tcp", ":0")
	responseChan := make(chan ConnectToResp)
	m.InUiConnectTo <- InUiConnectTo{Address: listener.Addr().String(), Resp: responseChan}

	response := <-responseChan

	if !response.Success {
		t.Error("Node did not connect")
	}
}

//func TestIncomingConnection(t *testing.T) {
//	dir := tmpDir()
//	defer cleanDir(dir, t)
//	userCmds := make(chan events.Event)
//	mockNet := test.MockNet{Dialed: false, AcceptsConnections: true}
//	n := mgr.New(userCmds, &mockNet, dir)
//	n.Start()
//
//	remoteNodeId := nodes.NewNodeId()
//	mockNet.Conn.SendMockBytes(validHello(remoteNodeId))
//
//	expected := validHello(n.GetId())
//	time.Sleep(time.Millisecond * 100)
//
//	result := readPayloadBytesWritten(&mockNet)
//	if !bytes.Equal(result, expected) {
//		t.Error("You didn't hello back!")
//	}
//}
//
//func TestSendNodeSyncAfterReceiveHello(t *testing.T) {
//	dir := tmpDir()
//	defer cleanDir(dir, t)
//	userCmds := make(chan events.Event)
//	tNet := test.MockNet{Dialed: false, AcceptsConnections: false}
//	n := mgr.New(userCmds, &tNet, dir)
//	remoteNodeId := nodes.NewNodeId()
//	remoteNodeAddress := "remoteAddress"
//	n.Start()
//
//	userCmds <- events.NewString(events.ConnectTo, remoteNodeAddress)
//
//	if !tNet.IsDialed() {
//		t.Error("Node did not connect")
//	}
//
//	time.Sleep(time.Millisecond * 50)
//
//	tNet.Conn.BytesWritten = make([]byte, 0)
//	tNet.Conn.SendMockBytes(validHello(remoteNodeId))
//
//	expected := CommandAndNodes{Command: 2, Nodes: set.NewSet[NodeInfo]()}
//	expected.Nodes.Add(NodeInfo{NodeId: string(remoteNodeId), Address: remoteNodeAddress})
//	expected.Nodes.Add(NodeInfo{NodeId: string(n.GetId()), Address: tNet.GetBinding()})
//
//	time.Sleep(time.Millisecond * 20)
//	payload := readPayloadBytesWritten(&tNet)
//	commandAndNodes, err := commandAndNodesFrom(payload)
//	if err != nil {
//		t.Error(err.Error())
//		return
//	}
//
//	if commandAndNodes.Command != expected.Command {
//		t.Error("Invalid command " + strconv.Itoa(int(commandAndNodes.Command)))
//	}
//
//	if !commandAndNodes.Nodes.Equal(&expected.Nodes) {
//		t.Error("Node set is not correct")
//	}
//}
//
//func TestSaveAndRead(t *testing.T) {
//	dir := tmpDir()
//	defer cleanDir(dir, t)
//	expected := "BlahBlaHereIsTheDataAndItsLong"
//	userCmds := make(chan events.Event)
//	tNet := test.MockNet{Dialed: false, AcceptsConnections: false}
//	n := mgr.New(userCmds, &tNet, dir)
//	n.Start()
//
//	tempDir, _ := os.MkdirTemp("", "*-test-save-mgr")
//	defer removeAll(tempDir, t)
//
//	userCmds <- events.NewString(events.AddData, expected)
//	time.Sleep(100 * time.Millisecond)
//	result := make(chan []byte)
//	h := []byte{0x29, 0x4a, 0xe5, 0x7d, 0xde, 0xda, 0x37, 0xca, 0xb7, 0x44, 0xf7, 0x9e, 0xad, 0xdc, 0x15, 0x82, 0xae, 0xdc, 0x0c, 0x3e, 0x51, 0x7c, 0xff, 0x81, 0x49, 0xa1, 0x8b, 0x85, 0x34, 0xd4, 0xc0, 0x45}
//	userCmds <- events.NewBytesWithResult(events.ReadData, h, result)
//	r := string(<-result)
//	if r != expected {
//		t.Error("Not equal result: ", r, " vs expected: ", expected)
//	}
//}
//
//func validHello(nodeId nodes.Id) []byte {
//
//	serializedHello := int8Serialized(1)
//	serializedNodeId := []byte(nodeId)
//	serializedNodeIdLen := intSerialized(len(serializedNodeId))
//
//	payload := append(append(serializedHello, serializedNodeIdLen...), serializedNodeId...)
//	serializedPayload := intSerialized(len(payload))
//
//	return append(serializedPayload, payload...)
//}
//
//func readPayloadBytesWritten(tnet *test.MockNet) []byte {
//	if len(tnet.Conn.BytesWritten) < 4 {
//		return make([]byte, 0)
//	}
//	payloadLen := binary.BigEndian.Uint32(tnet.Conn.BytesWritten)
//	returnVal := tnet.Conn.BytesWritten[:4+payloadLen]
//	tnet.Conn.BytesWritten = tnet.Conn.BytesWritten[4+payloadLen:]
//	return returnVal
//}
//
//type NodeInfo struct {
//	NodeId  string
//	Address string
//}
//
//type CommandAndNodes struct {
//	Command int8
//	Nodes   set.Set[NodeInfo]
//}
//
//func commandAndNodesFrom(data []byte) (*CommandAndNodes, error) {
//	length := binary.BigEndian.Uint32(data)
//	if int(length) != len(data)-4 {
//		return nil, errors.New("invalid length")
//	}
//	command := int8(data[4])
//	if command != 2 {
//		return nil, errors.New("not a SyncNodes")
//	}
//	nds := set.NewSet[NodeInfo]()
//
//	start := 5
//
//	for start < len(data) {
//		idLen := int(binary.BigEndian.Uint32(data[start:]))
//		start += 4
//		id := string(data[start : start+idLen])
//		start += idLen
//		addressLen := int(binary.BigEndian.Uint32(data[start:]))
//		start += 4
//		address := string(data[start : start+addressLen])
//		start += addressLen
//		nds.Add(NodeInfo{NodeId: id, Address: address})
//	}
//
//	return &CommandAndNodes{Command: command, Nodes: nds}, nil
//}
//
//func intSerialized(number int) []byte {
//	serializedInt := make([]byte, 4)
//	binary.BigEndian.PutUint32(serializedInt, uint32(number))
//	return serializedInt
//}
//
//func int8Serialized(number int8) []byte {
//	return []byte{byte(number)}
//}
//
//func nodeIdIsValid(mgr *mgr.Mgr) bool {
//	return len(mgr.GetId()) > 0
//}

func testListener(listener net.Listener, received chan []byte, toSend chan []byte) {
	//listener, err := net.Listen("tcp", ":0")
	//if err != nil {
	//	fmt.Println("Error listening:", err)
	//	return
	//}
	//defer listener.Close()

	for {
		AcceptReadSend(listener, received)
	}
}

func AcceptReadSend(listener net.Listener, received chan []byte) {
	conn, _ := listener.Accept()
	defer func(conn net.Conn) {
		_ = conn.Close()
	}(conn)
	readConnection(conn, received)
}

func readConnection(conn net.Conn, received chan []byte) {
	for {
		buf := make([]byte, 1024)
		n, err := conn.Read(buf)
		if err != nil {
			fmt.Println("Error reading:", err)
			return
		}
		received <- buf[:n]
	}
}

func removeAll(dir string, t *testing.T) {
	err := os.RemoveAll(dir)
	if err != nil {
		t.Errorf("Error [%v] deleting temp dir [%v]", err, dir)
	}
}

func tmpDir() store.Path {
	tempDir, _ := os.MkdirTemp("", "*-test")
	return store.NewPath(tempDir)
}

func cleanDir(path store.Path, t *testing.T) {
	removeAll(path.String(), t)
}
