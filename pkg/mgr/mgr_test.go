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
	"bytes"
	"tealfs/pkg/disk"
	"tealfs/pkg/model"
	"tealfs/pkg/set"
	"testing"

	"context"
)

func TestConnectToMgr(t *testing.T) {
	const expectedAddress = "some-address:123"

	m := NewWithChanSize(model.NewNodeId(), 0, "dummyAddress", "dummyPath", &disk.MockFileOps{}, model.Mirrored)
	err := m.Start()
	if err != nil {
		t.Error("Error starting", err)
		return
	}

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
	idsInFlight := set.NewSet[model.BlockId]()
	for range 100 {
		blockId := model.NewBlockId()
		ids = append(ids, blockId)
	}

	meCount := 0
	oneCount := 0
	twoCount := 0

	for _, blockId := range ids {
		idsInFlight.Add(blockId)
		m.WebdavMgrGets <- blockId

		select {
		case r := <-m.MgrDiskReads:
			meCount++
			if !idsInFlight.Exists(model.BlockId(r.Ptr.FileName)) {
				t.Error("expected to read to 1, got", r.Ptr.FileName)
				return
			}
			m.DiskMgrReads <- model.ReadResult{
				Ok:     true,
				Caller: m.NodeId,
				Data: model.RawData{
					Ptr: model.DiskPointer{
						NodeId:   m.NodeId,
						FileName: string(blockId),
					},
					Data: []byte{1, 2, 3},
				},
			}
		case s := <-m.MgrConnsSends:
			var nodeWithData model.NodeId
			if s.ConnId == expectedConnectionId1 {
				oneCount++
				nodeWithData = expectedNodeId1
			} else if s.ConnId == expectedConnectionId2 {
				twoCount++
				nodeWithData = expectedNodeId2
			} else {
				t.Error("expected to connect to", s.ConnId)
				return
			}
			m.ConnsMgrReceives <- model.ConnsMgrReceive{
				ConnId: s.ConnId,
				Payload: &model.ReadResult{
					Ok:     true,
					Caller: m.NodeId,
					Data: model.RawData{
						Ptr: model.DiskPointer{
							NodeId:   nodeWithData,
							FileName: string(blockId),
						},
						Data: []byte{1, 2, 3},
					},
				},
			}
		case w := <-m.MgrWebdavGets:
			if !idsInFlight.Exists(w.Block.Id) {
				t.Error("unexpected block id")
				return
			}
			idsInFlight.Remove(w.Block.Id)
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
		block := model.Block{
			Id:   model.NewBlockId(),
			Data: data,
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
			if !bytes.Equal(w.Data.Data, block.Data) {
				t.Error("expected the original block")
				return
			}
			m.DiskMgrWrites <- model.WriteResult{
				Ok:     true,
				Caller: m.NodeId,
				Ptr: model.DiskPointer{
					NodeId:   m.NodeId,
					FileName: string(block.Id),
				},
			}
			println("test")
		case s := <-m.MgrConnsSends:
			var nodeWithData model.NodeId
			if s.ConnId == expectedConnectionId1 {
				oneCount++
				nodeWithData = expectedNodeId1
			} else if s.ConnId == expectedConnectionId2 {
				twoCount++
				nodeWithData = expectedNodeId2
			} else {
				t.Error("expected to connect to", s.ConnId)
				return
			}
			m.ConnsMgrReceives <- model.ConnsMgrReceive{
				ConnId: s.ConnId,
				Payload: &model.WriteResult{
					Ok:     true,
					Caller: m.NodeId,
					Ptr: model.DiskPointer{
						NodeId:   nodeWithData,
						FileName: string(block.Id),
					},
				},
			}
			println("test")
		}
	}
	if meCount == 0 || oneCount == 0 || twoCount == 0 {
		t.Error("Expected everyone to fetch some data")
		return
	}
}

type connectedNode struct {
	address string
	conn    model.ConnId
	node    model.NodeId
}

func mgrWithConnectedNodes(nodes []connectedNode, t *testing.T) *Mgr {
	m := NewWithChanSize(model.NewNodeId(), 5, "dummyAddress", "dummyPath", &disk.MockFileOps{}, model.Mirrored)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	err := m.Start()
	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			case <-m.MgrWebdavIsPrimary:
			}
		}
	}()
	if err != nil {
		t.Error("Error starting", err)
	}
	var nodesInCluster []connectedNode

	for _, n := range nodes {
		// Send a message to Mgr indicating another
		// node has connected
		m.ConnsMgrStatuses <- model.NetConnectionStatus{
			Type: model.Connected,
			Id:   n.conn,
		}

		// Then Mgr should send an Iam payload to
		// the appropriate connection id with its
		// own node id
		expectedIam := <-m.MgrConnsSends
		payload := expectedIam.Payload
		switch p := payload.(type) {
		case *model.IAm:
			if p.NodeId != m.NodeId {
				t.Error("Unexpected nodeId")
				panic("Unexpected nodeId")
			}
			if expectedIam.ConnId != n.conn {
				t.Error("Unexpected connId")
				panic("Unexpected connId")
			}
		default:
			t.Error("Unexpected payload", p)
			panic("Unexpected payload")
		}

		// Send a message to Mgr indicating the newly
		// connected node has sent us an Iam payload
		iamPayload := model.IAm{
			NodeId:  n.node,
			Address: n.address,
		}
		m.ConnsMgrReceives <- model.ConnsMgrReceive{
			ConnId:  n.conn,
			Payload: &iamPayload,
		}

		<-m.MgrUiStatuses

		nodesInCluster = append(nodesInCluster, n)
		var payloadsFromMgr []model.MgrConnsSend

		for range nodesInCluster {
			payloadsFromMgr = append(payloadsFromMgr, <-m.MgrConnsSends)
		}

		expectedSyncNodes := expectedSyncNodesForCluster(nodesInCluster)
		syncNodesWeSent := assertAllPayloadsSyncNodes(t, payloadsFromMgr)

		if !cIdSnSliceEquals(expectedSyncNodes, syncNodesWeSent) {
			t.Error("Expected sync nodes to match", expectedSyncNodes, syncNodesWeSent)
			panic("Expected sync nodes to match")
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
			panic("Unexpected payload")
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
