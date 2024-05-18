package mgr

import (
	"tealfs/pkg/nodes"
	"tealfs/pkg/proto"
	"testing"
)

func TestConnectToRemoteNodeNew(t *testing.T) {
	const expectedAddress = "some-address:123"

	m := NewNew()
	m.Start()

	m.UiMgrConnectTos <- UiMgrConnectTo{
		Address: expectedAddress,
	}

	expectedMessage := <-m.MgrConnsConnectTos

	if expectedMessage.Address != expectedAddress {
		t.Error("Received address", expectedMessage.Address)
	}
}

func TestConnectToSuccess(t *testing.T) {
	const expectedAddress1 = "some-address:123"
	const expectedConnectionId1 = 1
	var expectedNodeId1 = nodes.NewNodeId()
	const expectedAddress2 = "some-address2:234"
	const expectedConnectionId2 = 2
	var expectedNodeId2 = nodes.NewNodeId()

	m := NewNew()
	m.Start()

	m.ConnsMgrStatuses <- ConnsMgrStatus{
		Type:          Connected,
		RemoteAddress: expectedAddress1,
		Id:            expectedConnectionId1,
	}

	expectedIam := <-m.MgrConnsSends
	payload := expectedIam.Payload
	switch p := payload.(type) {
	case *proto.IAm:
		if p.NodeId != m.nodeId {
			t.Error("Unexpected nodeId")
		}
		if expectedIam.ConnId != expectedConnectionId1 {
			t.Error("Unexpected connId")
		}
	}

	// Fixme: Understand what connAddress is for and test it or remove it
	m.ConnsMgrStatuses <- ConnsMgrStatus{
		Type:          Connected,
		RemoteAddress: expectedAddress2,
		Id:            expectedConnectionId2,
	}
	expectedIam2 := <-m.MgrConnsSends
	payload2 := expectedIam2.Payload
	switch p := payload2.(type) {
	case *proto.IAm:
		if p.NodeId != m.nodeId {
			t.Error("Unexpected nodeId")
		}
		if expectedIam2.ConnId != expectedConnectionId2 {
			t.Error("Unexpected connId")
		}
	default:
		t.Error("Unexpected payload", p)
	}

	iamPayload1 := proto.IAm{
		NodeId: expectedNodeId1,
	}

	m.ConnsMgrReceives <- ConnsMgrReceive{
		ConnId:  expectedConnectionId1,
		Payload: &iamPayload1,
	}

	expectedSendPayload1 := <-m.MgrConnsSends
	switch p := expectedSendPayload1.Payload.(type) {
	case *proto.SyncNodes:
		if p.Nodes.Len() != 1 {
			t.Error("Unexpected number of nodes", p.Nodes.Len())
		}
	default:
		t.Error("Unexpected payload", p)
	}

	iamPayload2 := proto.IAm{
		NodeId: expectedNodeId2,
	}

	m.ConnsMgrReceives <- ConnsMgrReceive{
		ConnId:  expectedConnectionId2,
		Payload: &iamPayload2,
	}

	expectedSendPayload2 := <-m.MgrConnsSends
	switch p := expectedSendPayload2.Payload.(type) {
	case *proto.SyncNodes:
		if p.Nodes.Len() != 2 {
			t.Error("Unexpected number of nodes", p.Nodes.Len())
		}
	default:
		t.Error("Unexpected payload", p)
	}
}
