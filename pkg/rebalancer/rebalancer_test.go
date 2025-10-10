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
	"tealfs/pkg/disk/dist"
	"tealfs/pkg/model"
	"tealfs/pkg/rebalancer"
	"tealfs/pkg/set"
	"testing"

	"github.com/google/uuid"
)

func TestRebalancer(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	inStart := make(chan rebalancer.BalanceReqId)
	inResp := make(chan rebalancer.ExistsResp)
	outExistsReq := make(chan rebalancer.ExistsReq)
	outSafeDelete := make(chan rebalancer.SafeDelete)
	onFsIds := set.NewMap[rebalancer.BalanceReqId, rebalancer.FilesystemBlockIdList]()
	nodeId := model.NewNodeId()
	outStoreItReq := make(chan rebalancer.StoreItReq)
	distributer := dist.NewMirrorDistributer(nodeId)

	r := rebalancer.Rebalancer{
		InStart:         inStart,
		InResp:          inResp,
		OutExistsReq:    outExistsReq,
		OutSafeDelete:   outSafeDelete,
		OnFilesystemIds: &onFsIds,
		NodeId:          nodeId,
		OutStoreItReq:   outStoreItReq,
		Distributer:     &distributer,
	}

	balanceReqId := rebalancer.BalanceReqId(uuid.NewString())

	idSet := set.NewSet[model.BlockId]()
	idSet.Add("block1")
	idSet.Add("block2")
	idSet.Add("block3")

	onFsIds.Add(balanceReqId, rebalancer.FilesystemBlockIdList{
		BlockIds:     set.NewSet[model.BlockId](),
		BalanceReqId: balanceReqId,
	})

	go r.Start(ctx)
	inStart <- rebalancer.BalanceReqId(balanceReqId)

	<-outExistsReq
	<-outExistsReq
	<-outExistsReq
}
