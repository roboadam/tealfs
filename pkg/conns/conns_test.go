// Copyright (C) 2025 Adam Hess
//
// This program is free software: you can redistribute it and/or modify it under
// the terms of the GNU Affero General Public License as published by the Free
// Software Foundation, version 3.
//
// This program is distributed in the hope that it will be useful, but WITHOUT
// ANY WARRANTY; without even the implied warranty of MERCHANTABILITY or
// FITNESS FOR A PARTICULAR PURPOSE. See the GNU Affero General Public License
// for more details.
//
// You should have received a copy of the GNU Affero General Public License
// along with this program. If not, see <https://www.gnu.org/licenses/>.

package conns

import (
	"context"
	"encoding/binary"
	"encoding/gob"
	"errors"
	"reflect"
	"tealfs/pkg/model"
	"testing"
)

func TestAcceptConn(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	_, status, _, _, _, provider := newConnsTest(ctx)
	provider.Listener.accept <- true
	s := <-status
	if s.Type != model.Connected {
		t.Error("Received address")
	}
}

func TestConnectToConns(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	_, outStatus, _, inConnectTo, _, _ := newConnsTest(ctx)
	const expectedAddress = "expectedAddress:1234"
	status := connectTo(expectedAddress, outStatus, inConnectTo)
	if status.Type != model.Connected {
		t.Error("Connection didn't work")
	}
}

func TestSendData(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	_, outStatus, _, inConnectTo, inSend, provider := newConnsTest(ctx)
	caller := model.NewNodeId()
	data := model.RawData{
		Ptr:  model.NewDiskPointer("destNode", "disk1", "blockId"),
		Data: []byte{1, 2, 3},
	}
	expected := model.NewWriteRequest(caller, data, "putBlockId")
	status := connectTo("address:123", outStatus, inConnectTo)
	inSend <- model.MgrConnsSend{
		ConnId:  status.Id,
		Payload: &expected,
	}

	// payload := collectPayload(provider.Conn.dataWritten)
	decoder := gob.NewDecoder(&provider.Conn.dataWritten)
	var payload model.Payload2
	err := decoder.Decode(payload)
	if err != nil {
		t.Error("Error decoding payload", err)
		return
	}

	switch p := payload.(type) {
	case *model.WriteRequest:
		if !reflect.DeepEqual(p, expected) {
			t.Error("WriteRequest not equal to expected value")
		}
	default:
		t.Error("Unexpected payload", p)
	}
}

func TestSendReadRequestNoConnected(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	_, _, outReceives, _, inSend, _ := newConnsTest(ctx)
	caller := model.NodeId("caller1")
	ptrs := []model.DiskPointer{
		model.NewDiskPointer("nodeId1", "disk1", "filename1"),
		model.NewDiskPointer("nodeId2", "disk2", "filename2"),
	}
	blockId := model.BlockId("blockId1")
	reqId := model.GetBlockId("reqId")
	request := model.NewReadRequest(caller, ptrs, blockId, reqId)

	inSend <- model.MgrConnsSend{
		ConnId:  0,
		Payload: &request,
	}
	outReceive := <-outReceives
	if outReceive.ConnId != 0 {
		t.Error("Expected ConnId to be 0")
		return
	}
	switch p := outReceive.Payload.(type) {
	case *model.ReadRequest:
		if p.BlockId() != request.BlockId() || p.Caller() != request.Caller() {
			t.Error("unexpected read request not equal")
			return
		}
		if len(p.Ptrs()) != 1 || p.Ptrs()[0] != request.Ptrs()[1] {
			t.Error("Expected ptrs to be equal")
			return
		}
	default:
		t.Error("Unexpected payload", p)
		return
	}
}

func TestSendReadRequestSendFailure(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	_, outStatus, outReceives, inConnectTo, inSend, connProvider := newConnsTest(ctx)
	status := connectTo("address:123", outStatus, inConnectTo)
	connProvider.Conn.WriteError = errors.New("some error writing")
	req := model.NewReadRequest(
		"caller1",
		[]model.DiskPointer{
			model.NewDiskPointer("nodeId1", "disk1", "filename1"),
			model.NewDiskPointer("nodeId2", "disk2", "filename2"),
		},
		"blockId1",
		"getBlockId1",
	)
	inSend <- model.MgrConnsSend{
		ConnId:  status.Id,
		Payload: &req,
	}
	outReceive := <-outReceives
	if outReceive.ConnId != 0 {
		t.Error("Expected ConnId to be 0")
		return
	}
	switch p := outReceive.Payload.(type) {
	case *model.ReadRequest:
		if p.BlockId() != req.BlockId() || p.Caller() != req.Caller() {
			t.Error("unexpected read request not equal")
			return
		}
		if len(p.Ptrs()) != 1 || p.Ptrs()[0] != req.Ptrs()[1] {
			t.Error("Expected ptrs to be equal")
			return
		}
	default:
		t.Error("Unexpected payload", p)
		return
	}
}

func TestConnectionError(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	_, outStatus, _, inConnectTo, _, provider := newConnsTest(ctx)
	provider.Conn.ReadError = errors.New("some error reading")
	firstStatus := connectTo("address:123", outStatus, inConnectTo)
	if firstStatus.Type != model.Connected {
		t.Error("expected to be connected")
		return
	}
	secondStatus := <-outStatus
	if secondStatus.Type != model.NotConnected {
		t.Error("Expected not connected status")
		return
	}
}

func collectPayload(channel chan []byte) []byte {
	data := <-channel
	size := binary.BigEndian.Uint32(data[:4])
	data = data[4:]
	for {
		if len(data) >= int(size) {
			return data
		}
		data = append(data, <-channel...)
	}
}

func TestGetData(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	_, outStatus, cmr, inConnectTo, _, provider := newConnsTest(ctx)
	status := connectTo("remoteAddress:123", outStatus, inConnectTo)
	disks := []model.DiskIdPath{{Id: "disk1", Path: "disk1path", Node: "node1"}}
	iam := model.NewIam("nodeId", disks, "localAddress:123", 1)
	encoder := gob.NewEncoder(&provider.Conn.dataToRead)
	err := encoder.Encode(iam)
	if err != nil {
		t.Error("Error encoding payload", err)
		return
	}

	result := <-cmr

	if result.ConnId != status.Id || !reflect.DeepEqual(result, iam) {
		t.Error("We didn't pass the message")
	}
}

func lenAsBytes(data []byte) []byte {
	size := uint32(len(data))
	buf := make([]byte, 4)
	binary.BigEndian.PutUint32(buf, size)
	return buf
}

func connectTo(address string, outStatus chan model.NetConnectionStatus, inConnectTo chan model.ConnectToNodeReq) model.NetConnectionStatus {
	inConnectTo <- model.ConnectToNodeReq{Address: address}
	return <-outStatus
}

func newConnsTest(ctx context.Context) (Conns, chan model.NetConnectionStatus, chan model.ConnsMgrReceive, chan model.ConnectToNodeReq, chan model.MgrConnsSend, *MockConnectionProvider) {
	outStatuses := make(chan model.NetConnectionStatus)
	outReceives := make(chan model.ConnsMgrReceive)
	inConnectTo := make(chan model.ConnectToNodeReq)
	inSends := make(chan model.MgrConnsSend)
	provider := NewMockConnectionProvider()
	c := NewConns(outStatuses, outReceives, inConnectTo, inSends, &provider, "dummyAddress:123", model.NewNodeId(), ctx)
	return c, outStatuses, outReceives, inConnectTo, inSends, &provider
}
