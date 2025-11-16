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
	"tealfs/pkg/model"
	"tealfs/pkg/rebalancer"
	"testing"
)

func TestExistsSender(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	inExistsReq := make(chan rebalancer.ExistsReq)
	inExistsResp := make(chan rebalancer.ExistsResp)
	outLocalExistsReq := make(chan rebalancer.ExistsReq)
	outLocalExistsResp := make(chan rebalancer.ExistsResp)
	outRemote := make(chan model.MgrConnsSend)
	nodeId := model.NewNodeId()
	remoteNodeId := model.NewNodeId()
	nodeConnMap := model.NewNodeConnectionMapper()
	nodeConnMap.SetAll(0, "someAddress:123", remoteNodeId)

	existsSender := rebalancer.MsgSender{
		InExistsReq:   inExistsReq,
		InExistsResp:  inExistsResp,
		OutExistsReq:  outLocalExistsReq,
		OutExistsResp: outLocalExistsResp,
		OutRemote:     outRemote,
		NodeId:        nodeId,
		NodeConnMap:   nodeConnMap,
	}
	go existsSender.Start(ctx)

	localExistsReq := rebalancer.ExistsReq{
		Caller:       nodeId,
		BalanceReqId: "balanceReqId1",
		ExistsId:     "existsId1",
		DestNodeId:   nodeId,
		DestDiskId:   "destDiskId1",
		DestBlockId:  "destBlockId1",
	}

	inExistsReq <- localExistsReq
	r1 := <-outLocalExistsReq
	if r1.DestBlockId != "destBlockId1" {
		t.Error("invalid block id")
	}

	remoteExistsReq := rebalancer.ExistsReq{
		Caller:       nodeId,
		BalanceReqId: "balanceReqId2",
		ExistsId:     "existsId2",
		DestNodeId:   remoteNodeId,
		DestDiskId:   "destDiskId2",
		DestBlockId:  "destBlockId2",
	}

	inExistsReq <- remoteExistsReq
	r2 := <-outRemote

	if r2.Payload.Type() != model.ExistsReqType {
		t.Error("invalid payload type")
	}

	if r2.Payload.(*rebalancer.ExistsReq).DestBlockId != "destBlockId2" {
		t.Error("invalid block id")
	}

	remoteExistsReqNotConnected := rebalancer.ExistsReq{
		Caller:       nodeId,
		BalanceReqId: "balanceReqId2",
		ExistsId:     "existsId2",
		DestNodeId:   "noRemoteConnected",
		DestDiskId:   "destDiskId2",
		DestBlockId:  "destBlockId2",
	}

	inExistsReq <- remoteExistsReqNotConnected

	select {
	case <-outLocalExistsReq:
		t.Error("unexpected message on outLocalExistsReq")
	case <-outRemote:
		t.Error("unexpected message on outRemoteExistsReq")
	default:
	}

	localExistsResp := rebalancer.ExistsResp{
		Req: rebalancer.ExistsReq{
			Caller:       nodeId,
			BalanceReqId: "balanceReq3",
			ExistsId:     "existsId3",
			DestNodeId:   "node3",
			DestDiskId:   "disk3",
			DestBlockId:  "block3",
		},
		Ok: true,
	}

	inExistsResp <- localExistsResp
	resp1 := <-outLocalExistsResp
	if resp1.Req.DestBlockId != "block3" {
		t.Error("invalid block id")
	}

	remoteExistsResp := rebalancer.ExistsResp{
		Req: rebalancer.ExistsReq{
			Caller:       remoteNodeId,
			BalanceReqId: "balance4",
			ExistsId:     "exists4",
			DestNodeId:   "node4",
			DestDiskId:   "disk4",
			DestBlockId:  "block4",
		},
		Ok: true,
	}

	inExistsResp <- remoteExistsResp
	mcs := <-outRemote

	if resp2, ok := mcs.Payload.(*rebalancer.ExistsResp); ok {
		if resp2.Req.DestBlockId != "block4" {
			t.Error("invalid block id")
		}
	} else {
		t.Error("invalid payload type")
	}

	remoteExistsRespNotConnected := rebalancer.ExistsResp{
		Req: rebalancer.ExistsReq{
			Caller:       "notConnectedNode",
			BalanceReqId: "balance5",
			ExistsId:     "exists5",
			DestNodeId:   "node5",
			DestDiskId:   "disk5",
			DestBlockId:  "block5",
		},
		Ok: true,
	}

	inExistsResp <- remoteExistsRespNotConnected

	select {
	case <-outLocalExistsResp:
		t.Error("unexpected message on outLocalExistsReq")
	case <-outRemote:
		t.Error("unexpected message on outRemoteExistsReq")
	default:
	}
}

func TestStoreItReq(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	inStoreItReq := make(chan rebalancer.StoreItReq)
	outStoreItReq := make(chan rebalancer.StoreItReq)
	outRemote := make(chan model.MgrConnsSend)

	sender := rebalancer.MsgSender{
		InStoreItReq:  inStoreItReq,
		OutStoreItReq: outStoreItReq,
		OutRemote:     outRemote,
		NodeId:        "localNodeId",
		NodeConnMap:   model.NewNodeConnectionMapper(),
	}
	go sender.Start(ctx)

	sender.NodeConnMap.SetAll(0, "someAddress1:123", "remoteNodeId")

	localReq := rebalancer.StoreItReq{
		NodeId: "localNodeId",
		DiskId: "localDiskId",
		Cmd: rebalancer.StoreItCmd{
			Caller:       "caller",
			BalanceReqId: "balanceReqId",
			StoreItId:    "storeItId",
			DestNodeId:   "destNodeId",
			DestDiskId:   "localDiskId",
			DestBlockId:  "destBlockId",
			ExistsReq: rebalancer.ExistsReq{
				Caller:       "caller2",
				BalanceReqId: "balanceRequestId",
				ExistsId:     "existsId",
				DestNodeId:   "destNodeId2",
				DestDiskId:   "destDiskId2",
				DestBlockId:  "destBlockId2",
			},
		},
	}

	inStoreItReq <- localReq
	<-outStoreItReq

	remoteReq := rebalancer.StoreItReq{
		NodeId: "remoteNodeId",
		DiskId: "localDiskId",
		Cmd: rebalancer.StoreItCmd{
			Caller:       "caller",
			BalanceReqId: "balanceReqId",
			StoreItId:    "storeItId",
			DestNodeId:   "destNodeId",
			DestDiskId:   "localDiskId",
			DestBlockId:  "destBlockId",
			ExistsReq: rebalancer.ExistsReq{
				Caller:       "caller2",
				BalanceReqId: "balanceRequestId",
				ExistsId:     "existsId",
				DestNodeId:   "destNodeId2",
				DestDiskId:   "destDiskId2",
				DestBlockId:  "destBlockId2",
			},
		},
	}

	inStoreItReq <- remoteReq
	mcs := <-outRemote

	if mcs.Payload.Type() != model.StoreItReqType {
		t.Error("invalid payload type")
	}

	if mcs.Payload.(*rebalancer.StoreItReq).NodeId != "remoteNodeId" {
		t.Error("invalid node id")
	}
}
