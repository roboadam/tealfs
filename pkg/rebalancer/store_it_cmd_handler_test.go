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
	"tealfs/pkg/set"
	"testing"
)

func TestStoreItCmd(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	inStoreItCmd := make(chan rebalancer.StoreItCmd, 1)
	inStoreItResp := make(chan rebalancer.StoreItResp, 1)
	outStoreItReq := make(chan rebalancer.StoreItReq, 1)
	outExistsResp := make(chan rebalancer.ExistsResp, 1)

	allDiskIds := set.NewSet[model.AddDiskReq]()
	localDisks := set.NewSet[disk.Disk]()

	handler := rebalancer.StoreItCmdHandler{
		InStoreItCmd:  inStoreItCmd,
		InStoreItResp: inStoreItResp,
		OutStoreItReq: outStoreItReq,
		OutExistsResp: outExistsResp,
		AllDiskIds:    &allDiskIds,
		LocalDisks:    &localDisks,
	}

	go handler.Start(ctx)

	inStoreItCmd <- rebalancer.StoreItCmd{
		Caller:       "callerNodeId",
		BalanceReqId: "balanceReqId",
		StoreItId:    "storeItId",
		DestNodeId:   ",
		DestDiskId:   "",
		DestBlockId:  "",
		ExistsReq:    rebalancer.ExistsReq{},
	}
}
