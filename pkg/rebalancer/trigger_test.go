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

package rebalancer

import (
	"context"
	"tealfs/pkg/model"
	"tealfs/pkg/set"
	"testing"
	"time"
)

func newTriggerTest(ctx context.Context) (*Trigger, chan struct{}, chan model.MgrConnsSend, chan ListOnDiskBlockIdsCmd) {
	inTrigger := make(chan struct{})
	outSends := make(chan model.MgrConnsSend)
	outLocalReq := make(chan ListOnDiskBlockIdsCmd)
	nodeId := model.NodeId("node1")
	mapper := model.NewNodeConnectionMapper()
	disks := set.NewSet[model.DiskInfo]()

	trigger := &Trigger{
		InTrigger:   inTrigger,
		OutSends:    outSends,
		OutLocalReq: outLocalReq,
		NodeId:      nodeId,
		Mapper:      mapper,
		Disks:       &disks,
	}

	go trigger.Start(ctx)

	return trigger, inTrigger, outSends, outLocalReq
}

func TestTrigger(t *testing.T) {
	t.Run("does not trigger when not primary", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		trigger, inTrigger, outSends, outLocalReq := newTriggerTest(ctx)

		// Add a disk for another node with a "greater" NodeId, so we are not primary.
		trigger.Disks.Add(model.DiskInfo{NodeId: "node2", DiskId: "disk2", Path: "/d2"})
		trigger.Mapper.SetAll(model.ConnId(1), "addr2", "node2")

		inTrigger <- struct{}{}

		select {
		case <-outLocalReq:
			t.Error("should not have sent a local request when not primary")
		case <-outSends:
			t.Error("should not have sent a remote request when not primary")
		default:
		}
	})

	t.Run("does not trigger when not all nodes are connected", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		trigger, inTrigger, outSends, outLocalReq := newTriggerTest(ctx)

		// We are node1. Add a disk for node0, so we are primary.
		trigger.Disks.Add(model.DiskInfo{NodeId: "node0", DiskId: "disk0", Path: "/d0"})
		// But we don't have a connection to node0 in the mapper yet, so AreAllConnected will be false.
		trigger.Mapper.SetNodeAddress("node0", "addr0") // This sets the address but not the connection.

		inTrigger <- struct{}{}

		select {
		case <-outLocalReq:
			t.Error("should not have sent a local request when not all nodes are connected")
		case <-outSends:
			t.Error("should not have sent a remote request when not all nodes are connected")
		default:
		}
	})

	t.Run("triggers when primary and all nodes are connected", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		trigger, inTrigger, outSends, outLocalReq := newTriggerTest(ctx)

		// We are node1. Add a disk for ourselves and another node. We are primary.
		trigger.Disks.Add(model.DiskInfo{NodeId: "node1", DiskId: "disk1", Path: "/d1"})
		trigger.Disks.Add(model.DiskInfo{NodeId: "node0", DiskId: "disk0", Path: "/d0"})

		// Set up connections to other nodes.
		trigger.Mapper.SetAll(model.ConnId(1), "addr0", "node0")

		inTrigger <- struct{}{}

		var localReq ListOnDiskBlockIdsCmd
		select {
		case localReq = <-outLocalReq:
		case <-time.After(100 * time.Millisecond):
			t.Fatal("timed out waiting for local request")
		}

		if localReq.Caller != trigger.NodeId {
			t.Errorf("unexpected caller in local request: got %s, want %s", localReq.Caller, trigger.NodeId)
		}

		remoteReq := <-outSends

		if remoteReq.ConnId != model.ConnId(1) {
			t.Errorf("unexpected ConnId for remote request: got %d, want %d", remoteReq.ConnId, 1)
		}

		payload, ok := remoteReq.Payload.(*ListOnDiskBlockIdsCmd)
		if !ok {
			t.Fatalf("unexpected payload type: %T", remoteReq.Payload)
		}

		if payload.BalanceReqId != localReq.BalanceReqId {
			t.Error("local and remote request IDs do not match")
		}
		if payload.Caller != trigger.NodeId {
			t.Errorf("unexpected caller in remote request: got %s, want %s", payload.Caller, trigger.NodeId)
		}

		select {
		case <-outSends:
			t.Error("unexpected extra remote request")
		default:
		}
	})

	t.Run("triggers when primary and there are no other nodes", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		trigger, inTrigger, outSends, outLocalReq := newTriggerTest(ctx)

		// We are node1. Add a disk for ourselves. We are primary.
		trigger.Disks.Add(model.DiskInfo{NodeId: "node1", DiskId: "disk1", Path: "/d1"})

		// No other nodes, so AreAllConnected should be true.
		inTrigger <- struct{}{}

		var localReq ListOnDiskBlockIdsCmd
		select {
		case localReq = <-outLocalReq:
			// good
		case <-time.After(100 * time.Millisecond):
			t.Fatal("timed out waiting for local request")
		}

		if localReq.Caller != trigger.NodeId {
			t.Errorf("unexpected caller in local request: got %s, want %s", localReq.Caller, trigger.NodeId)
		}

		// Check no remote messages
		select {
		case <-outSends:
			t.Error("unexpected remote request")
		case <-time.After(50 * time.Millisecond):
			// good
		}
	})
}
