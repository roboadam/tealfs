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
	"tealfs/pkg/disk"
	"tealfs/pkg/model"
	"tealfs/pkg/set"
	"testing"
)

func TestDeleteUnused(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	inRunCleanup := make(chan AllBlockId, 1)
	outRemoteDelete := make(chan model.MgrConnsSend, 10)
	outLocalDelete := make(chan disk.DeleteBlockId, 10)

	onDiskIds := set.NewMap[AllBlockId, AllBlockIdResp]()
	onFilesystemIds := set.NewMap[AllBlockId, AllBlockIdResp]()
	mapper := model.NewNodeConnectionMapper()
	nodeId := model.NodeId("node1")

	deleter := &DeleteUnused{
		InRunCleanup:    inRunCleanup,
		OutRemoteDelete: outRemoteDelete,
		OutLocalDelete:  outLocalDelete,
		OnDiskIds:       &onDiskIds,
		OnFilesystemIds: &onFilesystemIds,
		Mapper:          mapper,
		NodeId:          nodeId,
	}

	go deleter.Start(ctx)

	mapper.SetAll(model.ConnId(1), "addr2", "node2")
	mapper.SetAll(model.ConnId(2), "addr3", "node3")

	onDiskSet := set.NewSet[model.BlockId]()
	onDiskSet.Add("block1")
	onDiskSet.Add("block2")
	onDiskSet.Add("block3")
	onDiskSet.Add("block4")
	onDiskIds.Add("req1", AllBlockIdResp{BlockIds: onDiskSet})

	activeSet := set.NewSet[model.BlockId]()
	activeSet.Add("block1")
	activeSet.Add("block3")
	onFilesystemIds.Add("req1", AllBlockIdResp{BlockIds: activeSet})

	inRunCleanup <- "req1"

	deletedBlocks := set.NewSet[model.BlockId]()
	for range 2 {
		msg := <-outLocalDelete
		deletedBlocks.Add(msg.BlockId)
	}

	expectedDeletes := set.NewSet[model.BlockId]()
	expectedDeletes.Add("block2")
	expectedDeletes.Add("block4")

	if !deletedBlocks.Equal(&expectedDeletes) {
		t.Errorf("unexpected local deletes. got %v, want %v", deletedBlocks.GetValues(), expectedDeletes.GetValues())
	}

	remoteDeletes := make(map[model.ConnId]*set.Set[model.BlockId])
	numRemoteDeletes := 2 * 2 // 2 blocks * 2 connections
	for range numRemoteDeletes {
		msg := <-outRemoteDelete
		payload, _ := msg.Payload.(*disk.DeleteBlockId)
		if _, ok := remoteDeletes[msg.ConnId]; !ok {
			emptySet := set.NewSet[model.BlockId]()
			remoteDeletes[msg.ConnId] = &emptySet
		}
		remoteDeletes[msg.ConnId].Add(payload.BlockId)
	}

	connections := mapper.Connections()
	for _, connId := range connections.GetValues() {
		deletesForConn := remoteDeletes[connId]
		if !deletesForConn.Equal(&expectedDeletes) {
			t.Errorf("unexpected remote deletes for connId %d. got %v, want %v", connId, deletesForConn.GetValues(), expectedDeletes.GetValues())
		}
	}
}
