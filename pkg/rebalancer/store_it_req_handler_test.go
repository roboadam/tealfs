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
	"tealfs/pkg/rebalancer"
	"tealfs/pkg/set"
	"testing"
)

func TestStoreItReqHandler(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	inStoreItReq := make(chan rebalancer.StoreItReq)
	outStoreItResp := make(chan rebalancer.StoreItResp)

	fileOps := disk.MockFileOps{}
	d := disk.New(disk.NewPath("", &fileOps), "nodeId", "diskId", ctx)

	localDisks := set.NewSetFromSlice([]disk.Disk{d})

	handler := rebalancer.StoreItReqHandler{
		InStoreItReq:   inStoreItReq,
		OutStoreItResp: outStoreItResp,
		LocalDisks:     &localDisks,
	}
	go handler.Start(ctx)

	cmd := rebalancer.StoreItCmd{
		Caller:       "mainNodeId",
		BalanceReqId: "balanceReqId",
		StoreItId:    "storeItId",
		DestNodeId:   "nodeId",
		DestDiskId:   "diskId",
		DestBlockId:  "blockId",
	}
	inStoreItReq <- rebalancer.StoreItReq{
		Cmd:    cmd,
		NodeId: "nodeId",
		DiskId: "",
	}

}
