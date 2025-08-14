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

	for _, n := range nodes {
		nodeConnMapper.SetAll(n.conn, n.address, n.node)

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

	}

	return m, &fileOps, custodianCommands
}
