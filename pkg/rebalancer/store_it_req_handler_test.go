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
	"bytes"
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
		DestNodeId:   "destNodeId",
		DestDiskId:   "destDiskId",
		DestBlockId:  "blockId",
	}

	req := rebalancer.StoreItReq{
		Cmd:    cmd,
		NodeId: "nodeId",
		DiskId: "diskId",
	}
	inStoreItReq <- req
	resp := <-outStoreItResp

	if resp.Ok {
		t.Error("invalid response")
	}

	fileOps.WriteFile("blockId", []byte{1, 2, 3, 4, 5})

	inStoreItReq <- req
	resp = <-outStoreItResp

	if !resp.Ok {
		t.Error("invalid response")
	}

	if !bytes.Equal(resp.Block.Data, []byte{1, 2, 3, 4, 5}) {
		t.Error("unexpected data")
	}

}
