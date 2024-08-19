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
	"tealfs/pkg/model"
	"testing"
)

func TestConnectToMgr(t *testing.T) {
	const expectedAddress = "some-address:123"

	m := NewWithChanSize(0)
	m.Start()

	m.UiMgrConnectTos <- model.UiMgrConnectTo{
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
	var expectedNodeId1 = model.NewNodeId()
	const expectedAddress2 = "some-address2:234"
	const expectedConnectionId2 = 2
	var expectedNodeId2 = model.NewNodeId()

	mgrWithConnectedNodes([]connectedNode{
		{address: expectedAddress1, conn: expectedConnectionId1, node: expectedNodeId1},
		{address: expectedAddress2, conn: expectedConnectionId2, node: expectedNodeId2},
	}, t)
}

func TestReceiveSyncNodes(t *testing.T) {
	const sharedAddress = "some-address:123"
	const sharedConnectionId = 1
	var sharedNodeId = model.NewNodeId()
	const localAddress = "some-address2:234"
	const localConnectionId = 2
	var localNodeId = model.NewNodeId()
	const remoteAddress = "some-address3:345"
	var remoteNodeId = model.NewNodeId()

	m := mgrWithConnectedNodes([]connectedNode{
		{address: sharedAddress, conn: sharedConnectionId, node: sharedNodeId},
		{address: localAddress, conn: localConnectionId, node: localNodeId},
	}, t)

	sn := model.NewSyncNodes()
	sn.Nodes.Add(struct {
		Node    model.NodeId
		Address string
	}{Node: sharedNodeId, Address: sharedAddress})
	sn.Nodes.Add(struct {
		Node    model.NodeId
		Address string
	}{Node: remoteNodeId, Address: remoteAddress})
	m.ConnsMgrReceives <- model.ConnsMgrReceive{
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
	var expectedNodeId1 = model.NewNodeId()
	const expectedAddress2 = "some-address2:234"
	const expectedConnectionId2 = 2
	var expectedNodeId2 = model.NewNodeId()

	m := mgrWithConnectedNodes([]connectedNode{
		{address: expectedAddress1, conn: expectedConnectionId1, node: expectedNodeId1},
		{address: expectedAddress2, conn: expectedConnectionId2, node: expectedNodeId2},
	}, t)

	ids := []model.BlockId{}
	for range 100 {
		ids = append(ids, model.NewBlockId())
	}

	value := []byte("123")

	meCount := 0
	oneCount := 0
	twoCount := 0

	for _, id := range ids {
		m.ConnsMgrReceives <- model.ConnsMgrReceive{
			ConnId: expectedConnectionId1,
			Payload: &model.SaveData{
				Block: model.Block{
					Id:   id,
					Data: value,
					Hash: hash.ForData(value),
				},
			},
		}

		select {
		case w := <-m.MgrDiskWrites:
			meCount++
			if w.Id != id {
				t.Error("expected to write to 1, got", w.Id)
			}
		case s := <-m.MgrConnsSends:
			//Todo: s.Payload should be checked for the correct value
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

func TestReceiveDiskRead(t *testing.T) {
	const expectedAddress1 = "some-address:123"
	const expectedConnectionId1 = 1
	var expectedNodeId1 = model.NewNodeId()
	const expectedAddress2 = "some-address2:234"
	const expectedConnectionId2 = 2
	var expectedNodeId2 = model.NewNodeId()

	m := mgrWithConnectedNodes([]connectedNode{
		{address: expectedAddress1, conn: expectedConnectionId1, node: expectedNodeId1},
		{address: expectedAddress2, conn: expectedConnectionId2, node: expectedNodeId2},
	}, t)

	storeId1 := model.NewBlockId()
	data1 := []byte{0x00, 0x01, 0x02}
	hash1 := hash.ForData(data1)

	rr := model.ReadResult{
		Ok:      true,
		Message: "",
		Caller:  m.nodeId,
		Block: model.Block{
			Id:   storeId1,
			Data: data1,
			Hash: hash1,
		},
	}

	m.DiskMgrReads <- rr

	toWebdav := <-m.MgrWebdavGets

	if !rr.Equal(&toWebdav) {
		t.Errorf("rr didn't equal toWebdav")
	}

	rr2 := model.ReadResult{
		Ok:      true,
		Message: "",
		Caller:  expectedNodeId1,
		Block: model.Block{
			Id:   storeId1,
			Data: data1,
			Hash: hash1,
		},
	}

	m.DiskMgrReads <- rr2
	sent2 := <-m.MgrConnsSends

	expectedMCS2 := model.MgrConnsSend{
		ConnId:  expectedConnectionId1,
		Payload: &rr2,
	}

	if !sent2.Equal(&expectedMCS2) {
		t.Errorf("sent2 not equal expectedMCS2")
	}
}

func TestWebdavGet(t *testing.T) {
	const expectedAddress1 = "some-address:123"
	const expectedConnectionId1 = 1
	var expectedNodeId1 = model.NewNodeId()
	const expectedAddress2 = "some-address2:234"
	const expectedConnectionId2 = 2
	var expectedNodeId2 = model.NewNodeId()

	m := mgrWithConnectedNodes([]connectedNode{
		{address: expectedAddress1, conn: expectedConnectionId1, node: expectedNodeId1},
		{address: expectedAddress2, conn: expectedConnectionId2, node: expectedNodeId2},
	}, t)

	ids := []model.BlockId{}
	for range 100 {
		ids = append(ids, model.NewBlockId())
	}

	meCount := 0
	oneCount := 0
	twoCount := 0

	for _, id := range ids {
		m.WebdavMgrGets <- model.ReadRequest{
			Caller:  m.nodeId,
			BlockId: id,
		}

		select {
		case r := <-m.MgrDiskReads:
			meCount++
			if r.BlockId != id {
				t.Error("expected to read to 1, got", r.BlockId)
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

func TestWebdavPut(t *testing.T) {
	const expectedAddress1 = "some-address:123"
	const expectedConnectionId1 = 1
	var expectedNodeId1 = model.NewNodeId()
	const expectedAddress2 = "some-address2:234"
	const expectedConnectionId2 = 2
	var expectedNodeId2 = model.NewNodeId()

	m := mgrWithConnectedNodes([]connectedNode{
		{address: expectedAddress1, conn: expectedConnectionId1, node: expectedNodeId1},
		{address: expectedAddress2, conn: expectedConnectionId2, node: expectedNodeId2},
	}, t)

	blocks := []model.Block{}
	for i := range 100 {
		data := []byte{byte(i)}
		hash := hash.ForData(data)
		block := model.Block{
			Id:   model.NewBlockId(),
			Data: data,
			Hash: hash,
		}
		blocks = append(blocks, block)
	}

	meCount := 0
	oneCount := 0
	twoCount := 0

	for _, block := range blocks {
		m.WebdavMgrPuts <- block

		select {
		case w := <-m.MgrDiskWrites:
			meCount++
			if !w.Equal(&block) {
				t.Error("expected the origial block")
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
		t.Error("Expected everyone to fetch some data")
	}
}

type connectedNode struct {
	address string
	conn    model.ConnId
	node    model.NodeId
}

func mgrWithConnectedNodes(nodes []connectedNode, t *testing.T) Mgr {
	m := NewWithChanSize(0)
	m.Start()
	var nodesInCluster []connectedNode

	for _, n := range nodes {
		// Send a message to Mgr indicating another
		// node has connected
		m.ConnsMgrStatuses <- model.ConnectionStatus{
			Type:          model.Connected,
			RemoteAddress: n.address,
			Id:            n.conn,
		}

		// Then Mgr should send an Iam payload to
		// the appropriate connection id with its
		// own node id
		expectedIam := <-m.MgrConnsSends
		payload := expectedIam.Payload
		switch p := payload.(type) {
		case *model.IAm:
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
		iamPayload := model.IAm{
			NodeId: n.node,
		}
		m.ConnsMgrReceives <- model.ConnsMgrReceive{
			ConnId:  n.conn,
			Payload: &iamPayload,
		}

		nodesInCluster = append(nodesInCluster, n)
		var payloadsFromMgr []model.MgrConnsSend

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

func assertAllPayloadsSyncNodes(t *testing.T, mcs []model.MgrConnsSend) []connIdAndSyncNodes {
	var results []connIdAndSyncNodes
	for _, mc := range mcs {
		switch p := mc.Payload.(type) {
		case *model.SyncNodes:
			results = append(results, struct {
				ConnId  model.ConnId
				Payload model.SyncNodes
			}{ConnId: mc.ConnId, Payload: *p})
		default:
			t.Error("Unexpected payload", p)
		}
	}
	return results
}

type connIdAndSyncNodes struct {
	ConnId  model.ConnId
	Payload model.SyncNodes
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
	return a.Payload.Equal(&b.Payload)
}

func expectedSyncNodesForCluster(cluster []connectedNode) []connIdAndSyncNodes {
	var results []connIdAndSyncNodes

	sn := model.NewSyncNodes()
	for _, node := range cluster {
		sn.Nodes.Add(struct {
			Node    model.NodeId
			Address string
		}{Node: node.node, Address: node.address})
	}

	for _, node := range cluster {
		results = append(results, struct {
			ConnId  model.ConnId
			Payload model.SyncNodes
		}{ConnId: node.conn, Payload: sn})
	}
	return results
}
