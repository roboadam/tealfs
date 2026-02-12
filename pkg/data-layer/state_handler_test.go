// Copyright (C) 2026 Adam Hess
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

package datalayer_test

import (
	"context"
	datalayer "tealfs/pkg/data-layer"
	"tealfs/pkg/model"
	"testing"
)

func TestStateHandlerAsMain(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	outSave := make(chan datalayer.SaveRequest)
	outDelete := make(chan datalayer.DeleteRequest)
	outSends := make(chan model.SendPayloadMsg)

	mapper := model.NewNodeConnectionMapper()
	mapper.SetAll(0, "remoteNode1Address", "remoteNode1Id")
	mapper.SetAll(1, "remoteNode2Address", "remoteNode2Id")

	stateHandler := datalayer.StateHandler{
		OutSaveRequest:   outSave,
		OutDeleteRequest: outDelete,
		OutSends:         outSends,
		MainNodeId:       "nodeId",
		MyNodeId:         "nodeId",
		NodeConnMap:      mapper,
	}
	go stateHandler.Start(ctx)

	stateHandler.SetDiskSpace(datalayer.Dest{DiskId: "disk1Id", NodeId: "nodeId"}, 2)
	stateHandler.SetDiskSpace(datalayer.Dest{DiskId: "disk2Id", NodeId: "remoteNode1Id"}, 1)
	stateHandler.SetDiskSpace(datalayer.Dest{DiskId: "disk3Id", NodeId: "remoteNode2Id"}, 3)

	stateHandler.Saved("block1Id", datalayer.Dest{DiskId: "disk1Id", NodeId: "nodeId"})
	netMsg := <- outSends
	if netMsg.ConnId != 1 {
		t.Error("Sent to wrong connection")
	}
	if receivedSave, ok := netMsg.Payload.(datalayer.SaveRequest); ok {
		if receivedSave.BlockId != "block1Id" {
			t.Error("Invalid BlockId")
		}
		if len(receivedSave.From) != 1 {
			t.Error("Block starts off saved in only one place")
		}
		from := receivedSave.From[0]
		if from.NodeId != "nodeId" || from.DiskId != "disk1Id" {
			t.Error("Should be already saved on the local nodes only disk")
		}
		to :=  receivedSave.To
		if to.NodeId != "remoteNode2Id" || to.DiskId != "disk3Id" {
			t.Error("Should be saved to the biggest disk")
		}
	} else {
		t.Error("Expected SaveRequest")
	}

}
