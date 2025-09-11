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
)

func TestCollector(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	inDiskBlockIds := make(chan BlockIdList, 1)
	inFilesystemBlockIds := make(chan BlockIdList, 1)
	outFetchActiveIds := make(chan BalanceReqId, 1)
	outRunCleanup := make(chan BalanceReqId, 1)

	collector := Collector{
		InDiskBlockIds:       inDiskBlockIds,
		InFilesystemBlockIds: inFilesystemBlockIds,
		OutFetchActiveIds:    outFetchActiveIds,
		OutRunCleanup:        outRunCleanup,

		Mapper: model.NewNodeConnectionMapper(),
		NodeId: "node1",
	}
	go collector.Start(ctx)

	collector.Mapper.SetAll(model.ConnId(1), "addr2", "node2")
	collector.Mapper.SetAll(model.ConnId(1), "addr3", "node3")

	blockIds := set.NewSet[model.BlockId]()
	blockIds.Add("block1")
	blockIds.Add("block2")
	inDiskBlockIds <- BlockIdList{
		Caller:       "node2",
		BalanceReqId: "id1",
		BlockIds:     blockIds,
	}

	blockIds = set.NewSet[model.BlockId]()
	blockIds.Add("block2")
	blockIds.Add("block3")
	inDiskBlockIds <- BlockIdList{
		Caller:       "node1",
		BalanceReqId: "id1",
		BlockIds:     blockIds,
	}

	blockIds = set.NewSet[model.BlockId]()
	blockIds.Add("block3")
	blockIds.Add("block4")
	inDiskBlockIds <- BlockIdList{
		Caller:       "node3",
		BalanceReqId: "id1",
		BlockIds:     blockIds,
	}

	nextStep := <-outFetchActiveIds
	if nextStep != "id1" {
		t.Errorf("unexpected next step: got %s, want %s", nextStep, "id1")
	}

	blockIds = set.NewSet[model.BlockId]()
	blockIds.Add("block1")
	blockIds.Add("block2")
	blockIds.Add("block3")
	inFilesystemBlockIds <- BlockIdList{
		Caller:       "node1",
		BlockIds:     blockIds,
		BalanceReqId: "id1",
	}

	lastStep := <-outRunCleanup
	if lastStep != "id1" {
		t.Errorf("unexpected last step: got %s, want %s", lastStep, "id1")
	}

	expectedIds := set.NewSet[model.BlockId]()
	expectedIds.Add("block1")
	expectedIds.Add("block2")
	expectedIds.Add("block3")
	expectedIds.Add("block4")

	if resp, ok := collector.OnDiskIds.Get("id1"); ok {
		if !resp.BlockIds.Equal(&expectedIds) {
			t.Error("unexpected block IDs")
		}
	} else {
		t.Error("expected block IDs not found")
	}

	expectedIds.Remove("block4")
	if resp, ok := collector.OnFilesystemIds.Get("id1"); ok {
		if !resp.BlockIds.Equal(&expectedIds) {
			t.Error("unexpected block IDs")
		}
	} else {
		t.Error("expected block IDs not found")
	}
}
