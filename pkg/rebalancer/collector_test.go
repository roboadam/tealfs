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

	inDiskBlockIds := make(chan AllBlockIdResp)
	inFilesystemBlockIds := make(chan AllBlockIdResp)
	outFetchActiveIds := make(chan AllBlockId)
	outRunCleanup := make(chan AllBlockId)

	collector := Collector{
		InDiskBlockIds:       inDiskBlockIds,
		InFilesystemBlockIds: inFilesystemBlockIds,
		OutFetchActiveIds:    outFetchActiveIds,
		OutRunCleanup:        outRunCleanup,

		Mapper: model.NewNodeConnectionMapper(),
		NodeId: "node1",
	}
	collector.Start(ctx)

	collector.Mapper.SetAll(model.ConnId(1), "addr2", "node2")
	collector.Mapper.SetAll(model.ConnId(1), "addr3", "node3")

	blockIds := set.NewSet[model.BlockId]()
	blockIds.Add("block1")
	blockIds.Add("block2")
	inDiskBlockIds <- AllBlockIdResp{
		Caller:   "node1",
		Id:       "id1",
		BlockIds: blockIds,
	}
}
