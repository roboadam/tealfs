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
	"tealfs/pkg/hash"
	"tealfs/pkg/nodes"
	"tealfs/pkg/proto"
	"tealfs/pkg/store"
	"testing"
)

func TestConnectTo(t *testing.T) {
	const expectedAddress = "some-address:123"

	m := NewWithChanSize(0)
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

func TestReceiveSaveData(t *testing.T) {
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

	data := [][]byte{
		[]byte("000"),
		[]byte("001"),
		[]byte("002"),
		[]byte("003"),
		[]byte("004"),
		[]byte("005"),
		[]byte("006"),
		[]byte("007"),
		[]byte("008"),
		[]byte("009"),
		[]byte("010"),
		[]byte("011"),
		[]byte("012"),
		[]byte("013"),
		[]byte("014"),
		[]byte("015"),
		[]byte("016"),
		[]byte("017"),
		[]byte("018"),
		[]byte("019"),
		[]byte("020"),
	}

	meCount := 0
	oneCount := 0
	twoCount := 0

	for _, val := range data {
		m.ConnsMgrReceives <- ConnsMgrReceive{
			ConnId: expectedConnectionId1,
			Payload: &proto.SaveData{
				Block: store.Block{
					Id:       "1",
					Data:     val,
					Hash:     hash.ForData(val),
					Children: []store.Id{},
				},
			},
		}

		select {
		case w := <-m.MgrDiskWrites:
			meCount++
			if w.Id != "1" {
				t.Error("expected to write to 1, got", w.Id)
			}
		case s := <-m.MgrConnsSends:
			if s.ConnId == expectedConnectionId1 {
				oneCount++
			} else if s.ConnId == expectedConnectionId2 {
				twoCount++
			} else {
				t.Error("expected to connect to", s.ConnId)
			}
		}
	}
	if meCount == 0 || oneCount == 0 || twoCount == 0 {
		t.Error("Expected everyone to get some data")
	}
}

type connectedNode struct {
	address string
	conn    ConnId
	node    nodes.Id
}

func mgrWithConnectedNodes(nodes []connectedNode, t *testing.T) Mgr {
	m := NewWithChanSize(0)
	m.Start()
	var nodesInCluster []connectedNode

	for _, n := range nodes {
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

		for range nodesInCluster {
			payloadsFromMgr = append(payloadsFromMgr, <-m.MgrConnsSends)
		}

		expectedSyncNodes := expectedSyncNodesForCluster(nodesInCluster)
		syncNodesWeSent := assertAllPayloadsSyncNodes(t, payloadsFromMgr)

		if !cIdSnSliceEquals(expectedSyncNodes, syncNodesWeSent) {
			t.Error("Expected sync nodes to match", expectedSyncNodes, syncNodesWeSent)
		}
	}

	return m
}

func assertAllPayloadsSyncNodes(t *testing.T, mcs []MgrConnsSend) []connIdAndSyncNodes {
	var results []connIdAndSyncNodes
	for _, mc := range mcs {
		switch p := mc.Payload.(type) {
		case *proto.SyncNodes:
			results = append(results, struct {
				ConnId  ConnId
				Payload proto.SyncNodes
			}{ConnId: mc.ConnId, Payload: *p})
		default:
			t.Error("Unexpected payload", p)
		}
	}
	return results
}

type connIdAndSyncNodes struct {
	ConnId  ConnId
	Payload proto.SyncNodes
}

func cIdSnSliceEquals(a, b []connIdAndSyncNodes) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		oneEqual := false
		for j := range b {
			if cIdSnEquals(a[i], b[j]) {
				oneEqual = true
			}
		}
		if !oneEqual {
			return false
		}
	}
	return true
}

func cIdSnEquals(a, b connIdAndSyncNodes) bool {
	if a.ConnId != b.ConnId {
		return false
	}
	return a.Payload.Equals(&b.Payload)
}

func expectedSyncNodesForCluster(cluster []connectedNode) []connIdAndSyncNodes {
	var results []connIdAndSyncNodes

	sn := proto.NewSyncNodes()
	for _, node := range cluster {
		sn.Nodes.Add(struct {
			Node    nodes.Id
			Address string
		}{Node: node.node, Address: node.address})
	}

	for _, node := range cluster {
		results = append(results, struct {
			ConnId  ConnId
			Payload proto.SyncNodes
		}{ConnId: node.conn, Payload: sn})
	}
	return results
}
