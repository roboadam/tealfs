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

func TestRebalancerStart(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	inStart := make(chan rebalancer.BalanceReqId)
	inResp := make(chan rebalancer.ExistsResp)
	outExistsReq := make(chan rebalancer.ExistsReq)
	outSafeDelete := make(chan rebalancer.SafeDelete)
	onFsIds := set.NewMap[rebalancer.BalanceReqId, rebalancer.FilesystemBlockIdList]()
	nodeId := model.NewNodeId()
	outStoreItReq := make(chan rebalancer.StoreItCmd)
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
		Caller:       nodeId,
		BlockIds:     idSet,
		BalanceReqId: balanceReqId,
	})

	diskId1 := model.DiskId(uuid.NewString())
	diskId2 := model.DiskId(uuid.NewString())
	diskId3 := model.DiskId(uuid.NewString())
	distributer.SetWeight(nodeId, diskId1, 1)
	distributer.SetWeight(nodeId, diskId2, 1)
	distributer.SetWeight(nodeId, diskId3, 1)

	go r.Start(ctx)
	inStart <- rebalancer.BalanceReqId(balanceReqId)

	er1 := <-outExistsReq
	er2 := <-outExistsReq
	er3 := <-outExistsReq
	er4 := <-outExistsReq
	er5 := <-outExistsReq
	er6 := <-outExistsReq

	balanceReqIdCounter := set.Counter[rebalancer.BalanceReqId]{}
	balanceReqIdCounter.Tick(er1.BalanceReqId)
	balanceReqIdCounter.Tick(er2.BalanceReqId)
	balanceReqIdCounter.Tick(er3.BalanceReqId)
	balanceReqIdCounter.Tick(er4.BalanceReqId)
	balanceReqIdCounter.Tick(er5.BalanceReqId)
	balanceReqIdCounter.Tick(er6.BalanceReqId)

	if balanceReqIdCounter.Count(balanceReqId) != 6 {
		t.Error("invalid balance ids")
	}

	blockIdCounter := set.Counter[model.BlockId]{}
	blockIdCounter.Tick(er1.DestBlockId)
	blockIdCounter.Tick(er2.DestBlockId)
	blockIdCounter.Tick(er3.DestBlockId)
	blockIdCounter.Tick(er4.DestBlockId)
	blockIdCounter.Tick(er5.DestBlockId)
	blockIdCounter.Tick(er6.DestBlockId)

	if blockIdCounter.Count("block1") != 2 {
		t.Error("invalid block ids")
	}
	if blockIdCounter.Count("block2") != 2 {
		t.Error("invalid block ids")
	}
	if blockIdCounter.Count("block3") != 2 {
		t.Error("invalid block ids")
	}
}

func TestRebalancerAllExist(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	inStart := make(chan rebalancer.BalanceReqId)
	inResp := make(chan rebalancer.ExistsResp)
	outExistsReq := make(chan rebalancer.ExistsReq)
	outSafeDelete := make(chan rebalancer.SafeDelete)
	onFsIds := set.NewMap[rebalancer.BalanceReqId, rebalancer.FilesystemBlockIdList]()
	nodeId := model.NewNodeId()
	outStoreItReq := make(chan rebalancer.StoreItCmd)
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

	onFsIds.Add(balanceReqId, rebalancer.FilesystemBlockIdList{
		Caller:       nodeId,
		BlockIds:     idSet,
		BalanceReqId: balanceReqId,
	})

	diskId1 := model.DiskId(uuid.NewString())
	diskId2 := model.DiskId(uuid.NewString())
	distributer.SetWeight(nodeId, diskId1, 1)
	distributer.SetWeight(nodeId, diskId2, 1)

	go r.Start(ctx)
	inStart <- rebalancer.BalanceReqId(balanceReqId)

	er1 := <-outExistsReq
	er2 := <-outExistsReq

	eResp1 := rebalancer.ExistsResp{Req: er1, Ok: true}
	eResp2 := rebalancer.ExistsResp{Req: er2, Ok: true}

	inResp <- eResp1
	inResp <- eResp2

	sd1 := <-outSafeDelete

	if sd1.BlockId != "block1" {
		t.Error("invalid block id")
	}
}

func TestRebalancerNotExist(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	inStart := make(chan rebalancer.BalanceReqId)
	inResp := make(chan rebalancer.ExistsResp)
	outExistsReq := make(chan rebalancer.ExistsReq)
	outSafeDelete := make(chan rebalancer.SafeDelete)
	onFsIds := set.NewMap[rebalancer.BalanceReqId, rebalancer.FilesystemBlockIdList]()
	nodeId := model.NewNodeId()
	outStoreItReq := make(chan rebalancer.StoreItCmd)
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

	onFsIds.Add(balanceReqId, rebalancer.FilesystemBlockIdList{
		Caller:       nodeId,
		BlockIds:     idSet,
		BalanceReqId: balanceReqId,
	})

	diskId1 := model.DiskId(uuid.NewString())
	diskId2 := model.DiskId(uuid.NewString())
	distributer.SetWeight(nodeId, diskId1, 1)
	distributer.SetWeight(nodeId, diskId2, 1)

	go r.Start(ctx)
	inStart <- rebalancer.BalanceReqId(balanceReqId)

	er1 := <-outExistsReq
	er2 := <-outExistsReq

	eResp1 := rebalancer.ExistsResp{Req: er1, Ok: true}
	eResp2 := rebalancer.ExistsResp{Req: er2, Ok: false}

	inResp <- eResp1
	inResp <- eResp2

	storeIt := <-outStoreItReq

	if storeIt.DestBlockId != "block1" {
		t.Error("invalid block id")
	}
}
