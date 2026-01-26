// Copyright (C) 2026 Adam Hess
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
	"reflect"
	"tealfs/pkg/model"
	"tealfs/pkg/tnet"
	"testing"
	"time"
)

func TestAcceptConn(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	_, _, _, _, outSendIam, provider := newConnsTest(ctx)
	provider.Listener.accept <- true

	<-outSendIam
}

func TestConnectToConns(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	_, inConnectTo, _, _, sendIam, connProvider := newConnsTest(ctx)
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
	_, inConnectTo, inSend, _, sendIam, provider := newConnsTest(ctx)
	caller := model.NewNodeId()
	data := model.RawData{
		Ptr:  model.DiskPointer{NodeId: "destNode", Disk: "disk1", FileName: "blockId"},
		Data: []byte{1, 2, 3},
	}
	var expected model.Payload = &model.WriteRequest{Caller: caller, Data: data, ReqId: "putBlockId"}
	inConnectTo <- model.ConnectToNodeReq{Address: "address:123"}
	<-provider.DialedAddress
	connId := <-sendIam

	inSend <- model.SendPayloadMsg{
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

func newConnsTest(ctx context.Context) (
	*Conns,
	chan model.ConnectToNodeReq,
	chan model.SendPayloadMsg,
	chan model.IAm,
	chan model.ConnId,
	*MockConnectionProvider,
) {
	inConnectTo := make(chan model.ConnectToNodeReq, 1)
	inSends := make(chan model.SendPayloadMsg, 1)
	outIam := make(chan model.IAm, 1)
	outSendIam := make(chan model.ConnId, 1)
	provider := NewMockConnectionProvider()
	c := NewConns(inConnectTo, inSends, &provider, "dummyAddress:123", model.NewNodeId(), ctx)
	c.OutIam = outIam
	c.OutSendIam = outSendIam
	return c, inConnectTo, inSends, outIam, outSendIam, &provider
}
