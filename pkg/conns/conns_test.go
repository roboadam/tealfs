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
	"errors"
	"reflect"
	"tealfs/pkg/model"
	"tealfs/pkg/tnet"
	"testing"
	"time"
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
		Ptr:  model.DiskPointer{NodeId: "destNode", Disk: "disk1", FileName: "blockId"},
		Data: []byte{1, 2, 3},
	}
	var expected model.Payload = &model.WriteRequest{Caller: caller, Data: data, ReqId: "putBlockId"}
	status := connectTo("address:123", outStatus, inConnectTo)
	inSend <- model.MgrConnsSend{
		ConnId:  status.Id,
		Payload: expected,
	}
	time.Sleep(100 * time.Millisecond)

	rawNet := tnet.NewRawNet(&provider.Conn.dataWritten)
	payload, err := rawNet.ReadPayload()
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
		{NodeId: "nodeId1", Disk: "disk1", FileName: "filename1"},
		{NodeId: "nodeId2", Disk: "disk2", FileName: "filename2"},
	}
	blockId := model.BlockId("blockId1")
	reqId := model.GetBlockId("reqId")
	var request model.Payload = &model.ReadRequest{Caller: caller, Ptrs: ptrs, BlockId: blockId, ReqId: reqId}

	inSend <- model.MgrConnsSend{
		ConnId:  0,
		Payload: request,
	}
	outReceive := <-outReceives
	if outReceive.ConnId != 0 {
		t.Error("Expected ConnId to be 0")
		return
	}
	switch p := outReceive.Payload.(type) {
	case *model.ReadRequest:
		if p.BlockId != blockId || p.Caller != caller {
			t.Error("unexpected read request not equal")
			return
		}
		if len(p.Ptrs) != 1 || p.Ptrs[0] != ptrs[1] {
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
	var req model.Payload
	caller := model.NodeId("caller1")
	ptrs := []model.DiskPointer{
		{NodeId: "nodeId1", Disk: "disk1", FileName: "filename1"},
		{NodeId: "nodeId2", Disk: "disk2", FileName: "filename2"},
	}
	blockId := model.BlockId("blockId1")
	req = &model.ReadRequest{
		Caller:  caller,
		Ptrs:    ptrs,
		BlockId: blockId,
		ReqId:   "getBlockId1",
	}
	inSend <- model.MgrConnsSend{
		ConnId:  status.Id,
		Payload: req,
	}
	outReceive := <-outReceives
	if outReceive.ConnId != 0 {
		t.Error("Expected ConnId to be 0")
		return
	}
	switch p := outReceive.Payload.(type) {
	case *model.ReadRequest:
		if p.BlockId != blockId || p.Caller != caller {
			t.Error("unexpected read request not equal")
			return
		}
		if len(p.Ptrs) != 1 || p.Ptrs[0] != ptrs[1] {
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

func TestGetData(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	_, outStatus, cmr, inConnectTo, _, provider := newConnsTest(ctx)
	status := connectTo("remoteAddress:123", outStatus, inConnectTo)
	disks := []model.AddDiskReq{{Id: "disk1", Path: "disk1path", Node: "node1"}}

	buffer := ClosableBuffer{}
	rawNet := tnet.NewRawNet(&buffer)
	iam := model.NewIam("nodeId", disks, "localAddress:123")
	var payload model.Payload = &iam
	err := rawNet.SendPayload(payload)
	if err != nil {
		t.Error("Error sending payload", err)
		return
	}
	provider.Conn.dataToRead <- buffer

	result := <-cmr

	if result.ConnId != status.Id || !reflect.DeepEqual(result.Payload, payload) {
		t.Error("We didn't pass the message")
		return
	}
}

func connectTo(address string, outStatus chan model.NetConnectionStatus, inConnectTo chan model.ConnectToNodeReq) model.NetConnectionStatus {
	inConnectTo <- model.ConnectToNodeReq{Address: address}
	return <-outStatus
}

func newConnsTest(ctx context.Context) (*Conns, chan model.NetConnectionStatus, chan model.ConnsMgrReceive, chan model.ConnectToNodeReq, chan model.MgrConnsSend, *MockConnectionProvider) {
	outStatuses := make(chan model.NetConnectionStatus)
	outReceives := make(chan model.ConnsMgrReceive)
	inConnectTo := make(chan model.ConnectToNodeReq)
	inSends := make(chan model.MgrConnsSend)
	provider := NewMockConnectionProvider()
	c := NewConns(outStatuses, outReceives, inConnectTo, inSends, &provider, "dummyAddress:123", model.NewNodeId(), ctx)
	return c, outStatuses, outReceives, inConnectTo, inSends, &provider
}
