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

func TestExistsHandler(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	inExistsReq := make(chan rebalancer.ExistsReq, 1)
	outExistsResp := make(chan rebalancer.ExistsResp, 1)
	disks := set.NewSet[disk.Disk]()

	path := disk.NewPath("", &disk.MockFileOps{})
	d := disk.New(path, "nodeId", "diskId", ctx)
	disks.Add(d)

	existsHandler := rebalancer.ExistsHandler{
		InExistsReq:   inExistsReq,
		OutExistsResp: outExistsResp,

		Disks: &disks,
	}
	go existsHandler.Start(ctx)

	req := rebalancer.ExistsReq{
		Caller:       "caller",
		BalanceReqId: "balanceReqId",
		ExistsId:     "existsId",
		DestNodeId:   "nodeId",
		DestDiskId:   "diskId",
		DestBlockId:  "blockId",
	}

	inExistsReq <- req
	resp := <-outExistsResp

	if resp.Ok {
		t.Error("invalid response")
	}
}
