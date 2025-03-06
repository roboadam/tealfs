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
	"errors"
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
		Ptr: model.DiskPointer{
			NodeId:   "destNode",
			FileName: "blockId",
		},
		Data: []byte{1, 2, 3},
	}
	expected := model.NewWriteRequest(caller, data)
	status := connectTo("address:123", outStatus, inConnectTo)
	inSend <- model.MgrConnsSend{
		ConnId:  status.Id,
		Payload: &expected,
	}

	payload := collectPayload(provider.Conn.dataWritten)

	switch p := model.ToPayload(payload).(type) {
	case *model.WriteRequest:
		if !p.Equal(&expected) {
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
	// readRequest := model.ReadRequest{
	caller := model.NodeId("caller1")
	ptrs := []model.DiskPointer{
		{
			NodeId:   "nodeId1",
			FileName: "filename1",
		},
		{
			NodeId:   "nodeId2",
			FileName: "filename2",
		},
	}
	blockId := model.BlockId("blockId1")
	reqId := model.GetBlockId("reqid")
	// }
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
	readRequest := model.ReadRequest{
		Caller: "caller1",
		Ptrs: []model.DiskPointer{
			{
				NodeId:   "nodeId1",
				FileName: "filename1",
			},
			{
				NodeId:   "nodeId2",
				FileName: "filename2",
			},
		},
		BlockId: "blockId1",
	}
	inSend <- model.MgrConnsSend{
		ConnId:  status.Id,
		Payload: &readRequest,
	}
	outReceive := <-outReceives
	if outReceive.ConnId != 0 {
		t.Error("Expected ConnId to be 0")
		return
	}
	switch p := outReceive.Payload.(type) {
	case *model.ReadRequest:
		if p.BlockId != readRequest.BlockId || p.Caller != readRequest.Caller {
			t.Error("unexpected read request not equal")
			return
		}
		if len(p.Ptrs) != 1 || p.Ptrs[0] != readRequest.Ptrs[1] {
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
	payload := &model.IAm{
		NodeId:    "nodeId",
		Address:   "localAddress:123",
		FreeBytes: 1,
	}
	dataReceived := payload.ToBytes()
	length := lenAsBytes(dataReceived)
	provider.Conn.dataToRead <- length
	provider.Conn.dataToRead <- dataReceived

	result := <-cmr

	if result.ConnId != status.Id || !result.Payload.Equal(payload) {
		t.Error("We didn't pass the message")
	}
}

func lenAsBytes(data []byte) []byte {
	size := uint32(len(data))
	buf := make([]byte, 4)
	binary.BigEndian.PutUint32(buf, size)
	return buf
}

func connectTo(address string, outStatus chan model.NetConnectionStatus, inConnectTo chan model.MgrConnsConnectTo) model.NetConnectionStatus {
	inConnectTo <- model.MgrConnsConnectTo{Address: address}
	return <-outStatus
}

func newConnsTest(ctx context.Context) (Conns, chan model.NetConnectionStatus, chan model.ConnsMgrReceive, chan model.MgrConnsConnectTo, chan model.MgrConnsSend, *MockConnectionProvider) {
	outStatuses := make(chan model.NetConnectionStatus)
	outReceives := make(chan model.ConnsMgrReceive)
	inConnectTo := make(chan model.MgrConnsConnectTo)
	inSends := make(chan model.MgrConnsSend)
	provider := NewMockConnectionProvider()
	c := NewConns(outStatuses, outReceives, inConnectTo, inSends, &provider, "dummyAddress:123", model.NewNodeId(), ctx)
	return c, outStatuses, outReceives, inConnectTo, inSends, &provider
}
