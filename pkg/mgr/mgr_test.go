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

package mgr

import (
	"reflect"
	"sync/atomic"
	"tealfs/pkg/custodian"
	"tealfs/pkg/disk"
	"tealfs/pkg/model"
	"tealfs/pkg/set"
	"testing"
	"time"

	"context"

	"github.com/google/uuid"
)

func TestConnectToSuccess(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	const expectedAddress1 = "some-address:123"
	const expectedConnectionId1 = 1
	var expectedNodeId1 = model.NewNodeId()
	disks1 := []model.AddDiskReq{{DiskId: model.DiskId("disk1"), Path: "disk1path", NodeId: expectedNodeId1}}
	const expectedAddress2 = "some-address2:234"
	const expectedConnectionId2 = 2
	var expectedNodeId2 = model.NewNodeId()
	disks2 := []model.AddDiskReq{{DiskId: model.DiskId("disk2"), Path: "disk2path", NodeId: expectedNodeId2}}
	disks := []string{"disk"}

	_, _, _ = mgrWithConnectedNodes(
		ctx,
		[]connectedNode{
			{address: expectedAddress1, conn: expectedConnectionId1, node: expectedNodeId1, disks: disks1},
			{address: expectedAddress2, conn: expectedConnectionId2, node: expectedNodeId2, disks: disks2},
		}, 0, t, disks, make(chan<- model.ConnectToNodeReq))
}

func TestReceiveSyncNodes(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	const sharedAddress = "some-address:123"
	const sharedConnectionId = 1
	var sharedNodeId = model.NewNodeId()
	disks1 := []model.AddDiskReq{{DiskId: model.DiskId("disk1"), Path: "disk1path", NodeId: sharedNodeId}}
	const localAddress = "some-address2:234"
	const localConnectionId = 2
	var localNodeId = model.NewNodeId()
	disks2 := []model.AddDiskReq{{DiskId: model.DiskId("disk2"), Path: "disk2path", NodeId: localNodeId}}
	const remoteAddress = "some-address3:345"
	var remoteNodeId = model.NewNodeId()
	disks := []string{"disk"}
	connReqs := make(chan model.ConnectToNodeReq)

	m, _, _ := mgrWithConnectedNodes(ctx, []connectedNode{
		{address: sharedAddress, conn: sharedConnectionId, node: sharedNodeId, disks: disks1},
		{address: localAddress, conn: localConnectionId, node: localNodeId, disks: disks2},
	}, 0, t, disks, connReqs)

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

	expectedConnectTo := <-connReqs
	if expectedConnectTo.Address != remoteAddress {
		t.Error("expected to connect to", remoteAddress)
	}
}

func TestBroadcast(t *testing.T) {
	const expectedAddress1 = "some-address:123"
	const expectedConnectionId1 = 1
	var expectedNodeId1 = model.NewNodeId()
	disks1 := []model.AddDiskReq{{DiskId: "disk1", Path: "disk1path", NodeId: expectedNodeId1}}
	const expectedAddress2 = "some-address2:234"
	const expectedConnectionId2 = 2
	var expectedNodeId2 = model.NewNodeId()
	disks2 := []model.AddDiskReq{{DiskId: "disk2", Path: "disk2path", NodeId: expectedNodeId2}}
	maxNumberOfWritesInOnePass := 2
	paths := []string{"path1"}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	m, _, _ := mgrWithConnectedNodes(ctx, []connectedNode{
		{address: expectedAddress1, conn: expectedConnectionId1, node: expectedNodeId1, disks: disks1},
		{address: expectedAddress2, conn: expectedConnectionId2, node: expectedNodeId2, disks: disks2},
	}, maxNumberOfWritesInOnePass, t, paths, make(chan<- model.ConnectToNodeReq))

	testMsg := model.NewBroadcast([]byte{1, 2, 3})
	outMsgCounter := int32(0)

	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			case w := <-m.MgrConnsSends:
				if b, ok := w.Payload.(*model.Broadcast); ok {
					if reflect.DeepEqual(b, &testMsg) {
						atomic.AddInt32(&outMsgCounter, 1)
					}
				}
			}
		}
	}()

	m.WebdavMgrBroadcast <- model.NewBroadcast([]byte{1, 2, 3})
	time.Sleep(time.Millisecond * 500)
	ctr := atomic.LoadInt32(&outMsgCounter)
	if ctr != 2 {
		t.Error("Expected 2 messages to go out, got", outMsgCounter)
		return
	}

	msg := model.NewBroadcast([]byte{2, 3, 4})
	m.ConnsMgrReceives <- model.ConnsMgrReceive{
		ConnId:  expectedConnectionId1,
		Payload: &msg,
	}

	forwardedMsg := <-m.MgrWebdavBroadcast
	if !reflect.DeepEqual(forwardedMsg, msg) {
		t.Error("Wrong message was forwarded")
	}
}

type connectedNode struct {
	address string
	conn    model.ConnId
	node    model.NodeId
	disks   []model.AddDiskReq
}

func mgrWithConnectedNodes(ctx context.Context, nodes []connectedNode, chanSize int, t *testing.T, paths []string, connReqs chan<- model.ConnectToNodeReq) (*Mgr, *disk.MockFileOps, chan custodian.Command) {
	fileOps := disk.MockFileOps{}
	nodeConnMapper := model.NewNodeConnectionMapper()
	m := New(chanSize, "dummyAddress", "dummyPath", &fileOps, nodeConnMapper, ctx)
	m.ConnectToNodeReqs = connReqs
	custodianCommands := make(chan custodian.Command, chanSize)
	m.CustodianCommands = custodianCommands
	disks := set.NewSet[model.AddDiskReq]()
	m.AllDiskIds = &disks
	m.Start()

	for _, path := range paths {
		disks.Add(model.AddDiskReq{
			DiskId: model.DiskId(uuid.NewString()),
			Path:   path,
			NodeId: m.NodeId,
		})
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
		iamPayload := model.NewIam(n.node, n.disks, n.address)
		m.ConnsMgrReceives <- model.ConnsMgrReceive{
			ConnId:  n.conn,
			Payload: &iamPayload,
		}

		<-m.MgrUiConnectionStatuses
		for range n.disks {
			<-m.MgrUiDiskStatuses
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
			panic("Expected sync nodes to match")
		}
	}

	return m, &fileOps, custodianCommands
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
	return reflect.DeepEqual(a.Payload, b.Payload)
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
