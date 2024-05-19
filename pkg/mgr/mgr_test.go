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

	mgrWithConnectedNodes([]connectedNode{
		{address: expectedAddress1, conn: expectedConnectionId1, node: expectedNodeId1},
		{address: expectedAddress2, conn: expectedConnectionId2, node: expectedNodeId2},
	}, t)
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
		// Send a message to Mgr indicating another
		// node has connected
		m.ConnsMgrStatuses <- ConnsMgrStatus{
			Type:          Connected,
			RemoteAddress: n.address,
			Id:            n.conn,
		}

		// Then Mgr should send an Iam payload to
		// the appropriate connection id with its
		// own node id
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

		// Send a message to Mgr indicating the newly
		// connected node has sent us an Iam payload
		iamPayload := proto.IAm{
			NodeId: n.node,
		}
		m.ConnsMgrReceives <- ConnsMgrReceive{
			ConnId:  n.conn,
			Payload: &iamPayload,
		}

		// In response to the Iam, Mgr should reply
		// with a SyncNodes payload
		expectedSendPayload := <-m.MgrConnsSends
		switch p := expectedSendPayload.Payload.(type) {
		case *proto.SyncNodes:
			// The payload was not sent to the correct connection
			if expectedSendPayload.ConnId != n.conn {
				t.Error("Sent a SyncNodes to an unexpected connection")
			}

			// The length of the payload (in number of nodes synced) is incorrect
			if p.Nodes.Len() != i+1 {
				t.Error("Unexpected number of nodes", p.Nodes.Len())
			}

			for _, n := range p.Nodes.GetValues() {
				nodeIdIsExpected := false
				for _, expectedNode := range nodes {
					if n.Node == expectedNode.node {
						nodeIdIsExpected = true

						// The nodeId/address pair in the SyncNodes payload
						// did not match
						if n.Address != expectedNode.address {
							t.Error("Node id doesn't match address")
						}
					}
				}

				// At least one of the nodes in our SyncNodes payload has
				// a nodeId that is unexpected
				if !nodeIdIsExpected {
					t.Error("Unexpected node id", n)
				}
			}
		default:
			// We sent the wrong type of payload
			t.Error("Unexpected payload", p)
		}
	}

	return m
}
