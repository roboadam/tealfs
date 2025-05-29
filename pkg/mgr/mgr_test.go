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
	"fmt"
	"sync/atomic"
	"tealfs/pkg/chanutil"
	"tealfs/pkg/disk"
	"tealfs/pkg/model"
	"tealfs/pkg/set"
	"testing"
	"time"

	"context"

	"github.com/sirupsen/logrus"
)

func TestConnectToMgr(t *testing.T) {
	const expectedAddress = "some-address:123"
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	m := NewWithChanSize(
		0,
		"dummyAddress",
		"dummyPath",
		&disk.MockFileOps{},
		model.Mirrored,
		1,
		ctx,
	)

	m.UiMgrConnectTos <- model.UiMgrConnectTo{
		Address: expectedAddress,
	}

	expectedMessage := <-m.MgrConnsConnectTos

	if expectedMessage.Address != expectedAddress {
		t.Error("Received address", expectedMessage.Address)
	}
}

func TestConnectToSuccess(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	const expectedAddress1 = "some-address:123"
	const expectedConnectionId1 = 1
	var expectedNodeId1 = model.NewNodeId()
	disks1 := []model.DiskIdPath{{Id: model.DiskId("disk1"), Path: "disk1path", Node: expectedNodeId1}}
	const expectedAddress2 = "some-address2:234"
	const expectedConnectionId2 = 2
	var expectedNodeId2 = model.NewNodeId()
	disks2 := []model.DiskIdPath{{Id: model.DiskId("disk2"), Path: "disk2path", Node: expectedNodeId2}}
	disks := []string{"disk"}

	_, _ = mgrWithConnectedNodes(
		ctx,
		[]connectedNode{
			{address: expectedAddress1, conn: expectedConnectionId1, node: expectedNodeId1, disks: disks1},
			{address: expectedAddress2, conn: expectedConnectionId2, node: expectedNodeId2, disks: disks2},
		}, 0, t, disks)
}

func TestReceiveSyncNodes(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	const sharedAddress = "some-address:123"
	const sharedConnectionId = 1
	var sharedNodeId = model.NewNodeId()
	disks1 := []model.DiskIdPath{{Id: model.DiskId("disk1"), Path: "disk1path", Node: sharedNodeId}}
	const localAddress = "some-address2:234"
	const localConnectionId = 2
	var localNodeId = model.NewNodeId()
	disks2 := []model.DiskIdPath{{Id: model.DiskId("disk2"), Path: "disk2path", Node: localNodeId}}
	const remoteAddress = "some-address3:345"
	var remoteNodeId = model.NewNodeId()
	disks := []string{"disk"}

	m, _ := mgrWithConnectedNodes(ctx, []connectedNode{
		{address: sharedAddress, conn: sharedConnectionId, node: sharedNodeId, disks: disks1},
		{address: localAddress, conn: localConnectionId, node: localNodeId, disks: disks2},
	}, 0, t, disks)

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
	disks1 := []model.DiskIdPath{{Id: model.DiskId("disk1"), Path: "disk1path", Node: expectedNodeId1}}
	const expectedAddress2 = "some-address2:234"
	const expectedConnectionId2 = 2
	var expectedNodeId2 = model.NewNodeId()
	disks2 := []model.DiskIdPath{{Id: model.DiskId("disk2"), Path: "disk2path", Node: expectedNodeId2}}
	disks := []string{"disk"}
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	m, fileOps := mgrWithConnectedNodes(ctx, []connectedNode{
		{address: expectedAddress1, conn: expectedConnectionId1, node: expectedNodeId1, disks: disks1},
		{address: expectedAddress2, conn: expectedConnectionId2, node: expectedNodeId2, disks: disks2},
	}, 0, t, disks)

	ids := []model.BlockId{}
	for range 100 {
		blockId := model.NewBlockId()
		ids = append(ids, blockId)
	}

	oneCount := int32(0)
	twoCount := int32(0)

	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			case s := <-m.MgrConnsSends:
				switch readRequest := s.Payload.(type) {
				case *model.ReadRequest:
					if s.ConnId == expectedConnectionId1 {
						atomic.AddInt32(&oneCount, 1)
					} else if s.ConnId == expectedConnectionId2 {
						atomic.AddInt32(&twoCount, 1)
					}
					data := model.RawData{
						Ptr:  readRequest.Ptrs()[0],
						Data: []byte{1, 2, 3},
					}
					result := model.NewReadResultOk(
						readRequest.Caller(),
						readRequest.Ptrs()[1:],
						data,
						readRequest.GetBlockId(),
						readRequest.BlockId(),
					)
					m.ConnsMgrReceives <- model.ConnsMgrReceive{
						ConnId:  s.ConnId,
						Payload: &result,
					}
				}
			}
		}
	}()

	for _, blockId := range ids {
		m.WebdavMgrGets <- model.NewGetBlockReq(blockId)
		w := <-m.MgrWebdavGets
		if w.Block.Id != blockId {
			t.Error("Expected", blockId, "got", w.Block.Id)
		}
	}
	if fileOps.WriteCount == 0 || oneCount == 0 || twoCount == 0 {
		t.Error("Expected everyone to get some data")
		return
	}
}

func TestWebdavPut(t *testing.T) {
	paths := []string{"path1", "path2"}
	const expectedAddress1 = "some-address:1234"
	const expectedConnectionId1 = 1
	var expectedNodeId1 = model.NewNodeId()
	disks12 := []model.DiskIdPath{
		{Id: model.DiskId("disk1"), Path: "disk1path", Node: expectedNodeId1},
		{Id: model.DiskId("disk2"), Path: "disk2path", Node: expectedNodeId1},
	}
	const expectedAddress2 = "some-address2:234"
	const expectedConnectionId2 = 2
	var expectedNodeId2 = model.NewNodeId()
	disks34 := []model.DiskIdPath{
		{Id: model.DiskId("disk3"), Path: "disk3path", Node: expectedNodeId2},
		{Id: model.DiskId("disk4"), Path: "disk4path", Node: expectedNodeId2},
	}
	maxNumberOfWritesInOnePass := 2

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	m, fileOps := mgrWithConnectedNodes(ctx, []connectedNode{
		{address: expectedAddress1, conn: expectedConnectionId1, node: expectedNodeId1, disks: disks12},
		{address: expectedAddress2, conn: expectedConnectionId2, node: expectedNodeId2, disks: disks34},
	}, maxNumberOfWritesInOnePass, t, paths)

	blocks := []model.Block{}
	for i := range 5 {
		data := []byte{byte(i)}
		block := model.Block{
			Id:   model.NewBlockId(),
			Data: data,
		}
		blocks = append(blocks, block)
	}

	oneCount := int32(0)
	twoCount := int32(0)
	threeCount := int32(0)
	fourCount := int32(0)

	broadcasts := set.NewSet[model.BlockId]()
	cnt := 0

	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			case s := <-m.MgrConnsSends:
				switch request := s.Payload.(type) {
				case *model.WriteRequest:
					ptr := request.Data().Ptr
					if ptr.Disk() == disks12[0].Id {
						atomic.AddInt32(&oneCount, 1)
					} else if ptr.Disk() == disks12[1].Id {
						atomic.AddInt32(&twoCount, 1)
					} else if ptr.Disk() == disks34[0].Id {
						atomic.AddInt32(&threeCount, 1)
					} else if ptr.Disk() == disks34[1].Id {
						atomic.AddInt32(&fourCount, 1)
					}

					result := model.NewWriteResultOk(request.Data().Ptr, request.Caller(), request.ReqId())
					chanutil.Send(ctx, m.ConnsMgrReceives, model.ConnsMgrReceive{ConnId: s.ConnId, Payload: &result}, "remote")
				case *model.Broadcast:
					if request.Dest() == model.MgrDest {
						cmd := ToGlobalBlockListCommand(request.Msg())
						broadcasts.Add(cmd.BlockId)
						cnt++
						logrus.Infof("TEST CNT %d,%d - %s", cnt, broadcasts.Len(), cmd.BlockId)
					}
				}

			}
		}
	}()

	time.Sleep(time.Second * 2)

	for _, block := range blocks {
		m.WebdavMgrPuts <- model.NewPutBlockReq(block)
		<-m.MgrWebdavPuts
	}
	if fileOps.WriteCount == 0 || oneCount == 0 {
		t.Error("Expected everyone to fetch some data " + fmt.Sprintf("%d", fileOps.WriteCount))
		t.Error("Expected everyone to fetch some data " + fmt.Sprintf("%d", oneCount))
		return
	}

	if broadcasts.Len() != 5 {
		t.Errorf("Expected 100 broadcasts, got %d", broadcasts.Len())
		return
	}
}

func TestFileSystemBroadcast(t *testing.T) {
	const expectedAddress1 = "some-address:123"
	const expectedConnectionId1 = 1
	var expectedNodeId1 = model.NewNodeId()
	disks1 := []model.DiskIdPath{{Id: "disk1", Path: "disk1path", Node: expectedNodeId1}}
	const expectedAddress2 = "some-address2:234"
	const expectedConnectionId2 = 2
	var expectedNodeId2 = model.NewNodeId()
	disks2 := []model.DiskIdPath{{Id: "disk2", Path: "disk2path", Node: expectedNodeId2}}
	maxNumberOfWritesInOnePass := 2
	paths := []string{"path1"}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	m, _ := mgrWithConnectedNodes(ctx, []connectedNode{
		{address: expectedAddress1, conn: expectedConnectionId1, node: expectedNodeId1, disks: disks1},
		{address: expectedAddress2, conn: expectedConnectionId2, node: expectedNodeId2, disks: disks2},
	}, maxNumberOfWritesInOnePass, t, paths)

	testMsg := model.NewBroadcast([]byte{1, 2, 3}, model.FileSystemDest)
	outMsgCounter := 0

	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			case w := <-m.MgrConnsSends:
				if b, ok := w.Payload.(*model.Broadcast); ok {
					if b.Equal(&testMsg) {
						outMsgCounter++
					}
				}
			}
		}
	}()

	m.WebdavMgrBroadcast <- model.NewBroadcast([]byte{1, 2, 3}, model.FileSystemDest)
	time.Sleep(time.Millisecond * 500)
	if outMsgCounter != 2 {
		t.Error("Expected 2 messages to go out, got", outMsgCounter)
		return
	}

	msg := model.NewBroadcast([]byte{2, 3, 4}, model.FileSystemDest)
	m.ConnsMgrReceives <- model.ConnsMgrReceive{
		ConnId:  expectedConnectionId1,
		Payload: &msg,
	}

	forwardedMsg := <-m.MgrWebdavBroadcast
	if !forwardedMsg.Equal(&msg) {
		t.Error("Wrong message was forwarded")
	}
}

type connectedNode struct {
	address string
	conn    model.ConnId
	node    model.NodeId
	disks   []model.DiskIdPath
}

func mgrWithConnectedNodes(ctx context.Context, nodes []connectedNode, chanSize int, t *testing.T, paths []string) (*Mgr, *disk.MockFileOps) {
	fileOps := disk.MockFileOps{}
	m := NewWithChanSize(chanSize, "dummyAddress", "dummyPath", &fileOps, model.Mirrored, 1, ctx)

	for _, path := range paths {
		m.UiMgrDisk <- model.NewAddDiskReq(path, m.NodeId, 1)
		<-m.MgrUiDiskStatuses
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
			if p.Node() != m.NodeId {
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
		iamPayload := model.NewIam(n.node, n.disks, n.address, 1)
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

	return m, &fileOps
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
