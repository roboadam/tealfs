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
	_, _, _, _, _, outSendIam, provider := newConnsTest(ctx)
	provider.Listener.accept <- true

	<-outSendIam
}

func TestConnectToConns(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	_, _, inConnectTo, _, _, sendIam, connProvider := newConnsTest(ctx)
	const expectedAddress = "expectedAddress:1234"
	inConnectTo <- model.ConnectToNodeReq{Address: expectedAddress}

	actualAddress := <-connProvider.DialedAddress
	if actualAddress != expectedAddress {
		t.Error("Expected address to be dialed")
	}

	connId := <-sendIam
	if connId != 0 {
		t.Error("Expected connId to be 0")
	}
}

func TestSendData(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	_, _, inConnectTo, inSend, _, sendIam, provider := newConnsTest(ctx)
	caller := model.NewNodeId()
	data := model.RawData{
		Ptr:  model.DiskPointer{NodeId: "destNode", Disk: "disk1", FileName: "blockId"},
		Data: []byte{1, 2, 3},
	}
	var expected model.Payload = &model.WriteRequest{Caller: caller, Data: data, ReqId: "putBlockId"}
	inConnectTo <- model.ConnectToNodeReq{Address: "address:123"}
	<-provider.DialedAddress
	connId := <-sendIam

	inSend <- model.MgrConnsSend{
		ConnId:  connId,
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
	_, outReceives, _, inSend, _, _, _ := newConnsTest(ctx)
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
	_, outReceives, inConnectTo, inSend, _, sendIam, connProvider := newConnsTest(ctx)

	inConnectTo <- model.ConnectToNodeReq{Address: "address:123"}
	<-connProvider.DialedAddress
	connId := <-sendIam

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
		ConnId:  connId,
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

func newConnsTest(ctx context.Context) (
	*Conns,
	chan model.ConnsMgrReceive,
	chan model.ConnectToNodeReq,
	chan model.MgrConnsSend,
	chan model.IAm,
	chan model.ConnId,
	*MockConnectionProvider,
) {
	outReceives := make(chan model.ConnsMgrReceive, 1)
	inConnectTo := make(chan model.ConnectToNodeReq, 1)
	inSends := make(chan model.MgrConnsSend, 1)
	outIam := make(chan model.IAm, 1)
	outSendIam := make(chan model.ConnId, 1)
	provider := NewMockConnectionProvider()
	c := NewConns(outReceives, inConnectTo, inSends, &provider, "dummyAddress:123", model.NewNodeId(), ctx)
	c.OutIam = outIam
	c.OutSendIam = outSendIam
	return c, outReceives, inConnectTo, inSends, outIam, outSendIam, &provider
}
