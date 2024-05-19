// Copyright (C) 2024 Adam Hess
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

package mgr

import (
	"tealfs/pkg/nodes"
	"tealfs/pkg/proto"
	"tealfs/pkg/set"
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

func TestReceiveSyncNodes(t *testing.T) {
	const sharedAddress = "some-address:123"
	const sharedConnectionId = 1
	var sharedNodeId = nodes.NewNodeId()
	const localAddress = "some-address2:234"
	const localConnectionId = 2
	var localNodeId = nodes.NewNodeId()
	const remoteAddress = "some-address3:345"
	var remoteNodeId = nodes.NewNodeId()

	m := mgrWithConnectedNodes([]connectedNode{
		{address: sharedAddress, conn: sharedConnectionId, node: sharedNodeId},
		{address: localAddress, conn: localConnectionId, node: localNodeId},
	}, t)

	sn := proto.NewSyncNodes()
	sn.Nodes.Add(struct {
		Node    nodes.Id
		Address string
	}{Node: sharedNodeId, Address: sharedAddress})
	sn.Nodes.Add(struct {
		Node    nodes.Id
		Address string
	}{Node: remoteNodeId, Address: remoteAddress})
	m.ConnsMgrReceives <- ConnsMgrReceive{
		ConnId:  sharedConnectionId,
		Payload: &sn,
	}

	expectedConnectTo := <-m.MgrConnsConnectTos
	if expectedConnectTo.Address != remoteAddress {
		t.Error("expected to connect to", remoteAddress)
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
	var nodesInCluster []connectedNode

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

		nodesInCluster = append(nodesInCluster, n)
		var payloadsFromMgr []MgrConnsSend

		for _, _ = range nodesInCluster {
			payloadsFromMgr = append(payloadsFromMgr, <-m.MgrConnsSends)
		}

		syncNodesWeSent := assertAllPayloadsSyncNodes(t, payloadsFromMgr)
		connsWeSentTo := connsFromStruct(syncNodesWeSent)
		connsInCluster := connsFromConnectedNode(nodesInCluster)

		// The payloads were sent to all valid connections
		if !connsWeSentTo.Equal(&connsInCluster) {
			t.Error("Expected connsWeSent to equal connsInCluster")
		}
		// In response to the Iam, Mgr should reply
		// with a SyncNodes payload to all nodes it
		// has in the cluster
		for _, _ = range nodesInCluster {
			expectedSendPayload := <-m.MgrConnsSends
			switch p := expectedSendPayload.Payload.(type) {
			case *proto.SyncNodes:
				// The payload was not sent to a valid connection
				if expectedSendPayload.ConnId != n.conn {
					t.Error("Sent a SyncNodes to an unexpected connection got:", expectedSendPayload.ConnId, "instead of", n.conn)
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
	}

	return m
}

func connsFromConnectedNode(cns []connectedNode) set.Set[ConnId] {
	results := set.NewSet[ConnId]()
	for _, cn := range cns {
		results.Add(cn.conn)
	}
	return results
}

func connsFromStruct(mcs []struct {
	ConnId  ConnId
	Payload proto.Payload
}) set.Set[ConnId] {
	results := set.NewSet[ConnId]()
	for _, mc := range mcs {
		results.Add(mc.ConnId)
	}
	return results
}

func assertAllPayloadsSyncNodes(t *testing.T, mcs []MgrConnsSend) []struct {
	ConnId  ConnId
	Payload proto.Payload
} {
	var results []struct {
		ConnId  ConnId
		Payload proto.Payload
	}
	for _, mc := range mcs {
		switch p := mc.Payload.(type) {
		case *proto.SyncNodes:
			results = append(results, struct {
				ConnId  ConnId
				Payload proto.Payload
			}{ConnId: mc.ConnId, Payload: p})
		default:
			t.Error("Unexpected payload", p)
		}
	}
	return results
}
