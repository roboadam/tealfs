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

	inStoreItCmd := make(chan rebalancer.StoreItCmd)
	inStoreItResp := make(chan rebalancer.StoreItResp)
	outStoreItReq := make(chan rebalancer.StoreItReq)
	outExistsResp := make(chan rebalancer.ExistsResp)

	localDisks := set.NewSet[disk.Disk]()
	fileOps := disk.MockFileOps{}
	path := disk.NewPath("", &fileOps)
	localDisks.Add(disk.New(path, "localNode", "localDisk1", ctx))
	localDisks.Add(disk.New(path, "localNode", "localDisk2", ctx))

	allDiskIds := set.NewSet[model.DiskInfo]()
	allDiskIds.Add(model.DiskInfo{DiskId: "localDisk1", Path: "", NodeId: "localNode"})
	allDiskIds.Add(model.DiskInfo{DiskId: "localDisk2", Path: "", NodeId: "localNode"})
	allDiskIds.Add(model.DiskInfo{DiskId: "remoteDisk1", Path: "", NodeId: "remoteNode"})
	allDiskIds.Add(model.DiskInfo{DiskId: "remoteDisk2", Path: "", NodeId: "remoteNode"})

	expectedDisksToTry := set.NewSetFromSlice([]model.DiskId{"localDisk2", "remoteDisk1", "remoteDisk2"})

	handler := rebalancer.StoreItCmdHandler{
		InStoreItCmd:  inStoreItCmd,
		InStoreItResp: inStoreItResp,
		OutStoreItReq: outStoreItReq,
		OutExistsResp: outExistsResp,
		AllDiskIds:    &allDiskIds,
		LocalDisks:    &localDisks,
	}

	go handler.Start(ctx)

	cmd := rebalancer.StoreItCmd{
		Caller:       "mainNodeId",
		BalanceReqId: "balanceReqId",
		StoreItId:    "storeItId",
		DestNodeId:   "localNode",
		DestDiskId:   "localDisk1",
		DestBlockId:  "blockId",
		ExistsReq: rebalancer.ExistsReq{
			Caller:       "mainNodeId",
			BalanceReqId: "balanceReqId",
			ExistsId:     "existsId",
			DestNodeId:   "localNode",
			DestDiskId:   "localDisk1",
			DestBlockId:  "blockId",
		},
	}

	inStoreItCmd <- cmd

	disksTried := set.NewSet[model.DiskId]()
	req := <-outStoreItReq
	disksTried.Add(req.DiskId)

	inStoreItResp <- rebalancer.StoreItResp{Req: req, Ok: false, Msg: "failed"}

	req = <-outStoreItReq
	disksTried.Add(req.DiskId)

	inStoreItResp <- rebalancer.StoreItResp{Req: req, Ok: false, Msg: "failed"}
	req = <-outStoreItReq
	disksTried.Add(req.DiskId)

	inStoreItResp <- rebalancer.StoreItResp{Req: req, Ok: false, Msg: "failed"}
	select {
	case <-outStoreItReq:
		t.Error("unexpected message on outStoreItReq")
	default:
	}

	if !expectedDisksToTry.Equal(&disksTried) {
		t.Error("unexpected disks tried")
	}
}

func TestStoreItCmdSuccess(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	inStoreItCmd := make(chan rebalancer.StoreItCmd)
	inStoreItResp := make(chan rebalancer.StoreItResp)
	outStoreItReq := make(chan rebalancer.StoreItReq)
	outExistsResp := make(chan rebalancer.ExistsResp)

	localDisks := set.NewSet[disk.Disk]()
	fileOps := disk.MockFileOps{}
	path := disk.NewPath("", &fileOps)
	localDisks.Add(disk.New(path, "localNode", "localDisk1", ctx))
	localDisks.Add(disk.New(path, "localNode", "localDisk2", ctx))

	allDiskIds := set.NewSet[model.DiskInfo]()
	allDiskIds.Add(model.DiskInfo{DiskId: "localDisk1", Path: "", NodeId: "localNode"})
	allDiskIds.Add(model.DiskInfo{DiskId: "localDisk2", Path: "", NodeId: "localNode"})
	allDiskIds.Add(model.DiskInfo{DiskId: "remoteDisk1", Path: "", NodeId: "remoteNode"})
	allDiskIds.Add(model.DiskInfo{DiskId: "remoteDisk2", Path: "", NodeId: "remoteNode"})

	handler := rebalancer.StoreItCmdHandler{
		InStoreItCmd:  inStoreItCmd,
		InStoreItResp: inStoreItResp,
		OutStoreItReq: outStoreItReq,
		OutExistsResp: outExistsResp,
		AllDiskIds:    &allDiskIds,
		LocalDisks:    &localDisks,
	}

	go handler.Start(ctx)

	cmd := rebalancer.StoreItCmd{
		Caller:       "mainNodeId",
		BalanceReqId: "balanceReqId",
		StoreItId:    "storeItId",
		DestNodeId:   "localNode",
		DestDiskId:   "localDisk1",
		DestBlockId:  "blockId",
		ExistsReq: rebalancer.ExistsReq{
			Caller:       "mainNodeId",
			BalanceReqId: "balanceReqId",
			ExistsId:     "existsId",
			DestNodeId:   "localNode",
			DestDiskId:   "localDisk1",
			DestBlockId:  "blockId",
		},
	}

	inStoreItCmd <- cmd

	disksTried := set.NewSet[model.DiskId]()
	req := <-outStoreItReq
	disksTried.Add(req.DiskId)

	block := model.Block{
		Data: []byte("blockData"),
		Id:   "blockId",
	}
	inStoreItResp <- rebalancer.StoreItResp{Req: req, Ok: true, Block: block}

	existsResp := <-outExistsResp
	if existsResp.Req.ExistsId != "existsId" {
		t.Error("invalid exists id")
	}

	data, err := fileOps.ReadFile("blockId")
	if err != nil {
		t.Error("error verifying data")
	}

	if string(data) != "blockData" {
		t.Error("invalid data")
	}
}
