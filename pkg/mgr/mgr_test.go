package mgr

import (
	"tealfs/pkg/nodes"
	"tealfs/pkg/proto"
	"testing"
)

func TestConnectTo(t *testing.T) {
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

	m := mgrWithConnectedNodes([]connectedNode{
		{address: expectedAddress1, conn: expectedConnectionId1, node: expectedNodeId1},
		{address: expectedAddress2, conn: expectedConnectionId2, node: expectedNodeId2},
	}, t)

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
		for _, n := range p.Nodes.GetValues() {
			if n.Node != expectedNodeId1 && n.Node != expectedNodeId2 {
				t.Error("Unexpected node", n)
			}
			if n.Node == expectedNodeId1 && n.Address != expectedAddress1 {
				t.Error("Node id doesn't match address")
			}
			if n.Node == expectedNodeId2 && n.Address != expectedAddress2 {
				t.Error("Node id doesn't match address")
			}
		}
	default:
		t.Error("Unexpected payload", p)
	}
}

type connectedNode struct {
	address string
	conn    ConnId
	node    nodes.Id
}

func mgrWithConnectedNodes(nodes []connectedNode, t *testing.T) Mgr {
	m := NewNew()
	m.Start()

	for i, n := range nodes {
		m.ConnsMgrStatuses <- ConnsMgrStatus{
			Type:          Connected,
			RemoteAddress: n.address,
			Id:            n.conn,
		}

		expectedIam := <-m.MgrConnsSends
		payload := expectedIam.Payload
		switch p := payload.(type) {
		case *proto.IAm:
			if p.NodeId != m.nodeId {
				t.Error("Unexpected nodeId")
			}
			if expectedIam.ConnId != n.conn {
				t.Error("Unexpected connId")
			}
		default:
			t.Error("Unexpected payload", p)
		}

		iamPayload := proto.IAm{
			NodeId: n.node,
		}

		m.ConnsMgrReceives <- ConnsMgrReceive{
			ConnId:  n.conn,
			Payload: &iamPayload,
		}

		expectedSendPayload := <-m.MgrConnsSends
		switch p := expectedSendPayload.Payload.(type) {
		case *proto.SyncNodes:
			if p.Nodes.Len() != i+1 {
				t.Error("Unexpected number of nodes", p.Nodes.Len())
			}
			for _, n := range p.Nodes.GetValues() {
				if n.Node != expectedNodeId1 && n.Node != expectedNodeId2 {
					t.Error("Unexpected node", n)
				}
				if n.Node == expectedNodeId1 && n.Address != expectedAddress1 {
					t.Error("Node id doesn't match address")
				}
				if n.Node == expectedNodeId2 && n.Address != expectedAddress2 {
					t.Error("Node id doesn't match address")
				}
			}
		default:
			t.Error("Unexpected payload", p)
		}
	}

	return m
}
