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

package rebalancer_test

import (
	"context"
	"tealfs/pkg/disk"
	"tealfs/pkg/model"
	"tealfs/pkg/rebalancer"
	"tealfs/pkg/webdav"
	"testing"
)

func TestActiveBlockIdLister(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	inFetchIds := make(chan rebalancer.ListOnDiskBlockIdsCmd)
	outLocalResults := make(chan rebalancer.FilesystemBlockIdList)

	inBroadcast := make(chan webdav.FileBroadcast)
	fileOps := disk.MockFileOps{}
	outSends := make(chan model.MgrConnsSend, 1)
	mapper := *model.NewNodeConnectionMapper()

	fileSystem := webdav.NewFileSystem("nodeId", inBroadcast, &fileOps, "indexPath", 1, outSends, &mapper, ctx)

	lister := rebalancer.ActiveBlockIdLister{
		InFetchIds:      inFetchIds,
		OutLocalResults: outLocalResults,
		FileSystem:      &fileSystem,
	}
	go lister.Start(ctx)

	inFetchIds <- rebalancer.ListOnDiskBlockIdsCmd{
		Caller:       "caller",
		BalanceReqId: "balanceReqId",
	}
	<-outLocalResults
}
