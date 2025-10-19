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
	outLocalExistsReq := make(chan rebalancer.ExistsReq)
	outRemoteExistsReq := make(chan model.MgrConnsSend)
	nodeId := model.NewNodeId()
	remoteNodeId := model.NewNodeId()
	nodeConnMap := model.NewNodeConnectionMapper()
	nodeConnMap.SetAll(0, "someAddress:123", remoteNodeId)

	existsSender := rebalancer.ExistsSender{
		InExistsReq:        inExistsReq,
		OutLocalExistsReq:  outLocalExistsReq,
		OutRemoteExistsReq: outRemoteExistsReq,
		NodeId:             nodeId,
		NodeConnMap:        nodeConnMap,
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
	r2 := <-outRemoteExistsReq

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
	case <-outRemoteExistsReq:
		t.Error("unexpected message on outRemoteExistsReq")
	default:
	}
}
