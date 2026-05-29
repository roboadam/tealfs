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
	"tealfs/pkg/datalayer"
	"tealfs/pkg/model"
	"tealfs/pkg/test"
	"testing"
)

func TestStateHandlerAsMain(t *testing.T) {
	outPayload := make(chan model.Payload2, 20)

	remoteNode1Id := test.NonMainNode1
	remoteNode2Id := test.NonMainNode2
	myNodeId := test.MainNodeId

	mapper := model.NewNodeConnectionMapper()
	mapper.SetAll(0, "remoteNode1Address", remoteNode1Id)
	mapper.SetAll(1, "remoteNode2Address", remoteNode2Id)

	stateHandler := datalayer.StateHandler{
		MyNodeId:    myNodeId,
		NodeConnMap: mapper,
		OutPayload:  outPayload,
	}
	stateHandler.Start()

	stateHandler.SetDiskSpace(model.NodeDisk{DiskId: "disk1Id", NodeId: myNodeId}, 2)
	stateHandler.SetDiskSpace(model.NodeDisk{DiskId: "disk2Id", NodeId: remoteNode1Id}, 1)
	stateHandler.SetDiskSpace(model.NodeDisk{DiskId: "disk3Id", NodeId: remoteNode2Id}, 3)

	stateHandler.Saved("block1Id", model.NodeDisk{DiskId: "disk1Id", NodeId: myNodeId})
	payload := <-outPayload
	if receivedSave, ok := payload.(*datalayer.SaveRequest); ok {
		if receivedSave.BlockId != "block1Id" {
			t.Error("Invalid BlockId")
		}
		if len(receivedSave.From) != 1 {
			t.Error("Block starts off saved in only one place")
		}
		from := receivedSave.From[0]
		if from.NodeId != myNodeId || from.DiskId != "disk1Id" {
			t.Error("Should be already saved on the local nodes only disk")
		}
		to := receivedSave.To
		if to.NodeId != remoteNode2Id || to.DiskId != "disk3Id" {
			t.Error("Should be saved to the biggest disk")
		}
	}

	stateHandler.Saved("block2Id", model.NodeDisk{DiskId: "disk3Id", NodeId: remoteNode2Id})
	receivedPayload := <-outPayload
	if receivedSave, ok := receivedPayload.(*datalayer.SaveRequest); ok {
		if receivedSave.BlockId != "block2Id" {
			t.Error("Invalid BlockId")
		}
		if len(receivedSave.From) != 1 {
			t.Error("Block starts off saved in only one place")
		}
		from := receivedSave.From[0]
		if from.NodeId != remoteNode2Id || from.DiskId != "disk3Id" {
			t.Error("Should be already saved on the local nodes only disk")
		}
		to := receivedSave.To
		if to.NodeId != myNodeId || to.DiskId != "disk1Id" {
			t.Error("Should be saved to the biggest disk")
		}
	} else {
		t.Error("wrong type")
	}

	stateHandler.SetDiskSpace(model.NodeDisk{DiskId: "disk1Id", NodeId: myNodeId}, 1)
	stateHandler.SetDiskSpace(model.NodeDisk{DiskId: "disk2Id", NodeId: remoteNode1Id}, 2)
	stateHandler.SetDiskSpace(model.NodeDisk{DiskId: "disk3Id", NodeId: remoteNode2Id}, 3)
	stateHandler.Saved("block3id", model.NodeDisk{DiskId: "disk1Id", NodeId: myNodeId})

	payload = <-outPayload
	s1, ok := payload.(*datalayer.SaveRequest)
	if !ok {
		t.Fatal("invalid type")
	}
	stateHandler.Saved(s1.BlockId, s1.To)

	payload = <-outPayload
	s2, ok := payload.(*datalayer.SaveRequest)
	if !ok {
		t.Fatal("invalid type")
	}
	stateHandler.Saved(s2.BlockId, s2.To)

	payload = <-outPayload
	if del, ok := payload.(*datalayer.DeleteRequest); ok {
		stateHandler.Deleted(del.BlockId, del.Dest)
	} else {
		t.Fatal("Wrong type")
	}

	stateHandler.Deleted(s1.BlockId, s1.To)
	payload = <-outPayload
	if _, ok := payload.(*datalayer.DeleteRequest); !ok {
		t.Fatal("wrong type")
	}
	payload = <-outPayload
	if s3, ok := payload.(*datalayer.SaveRequest); ok {
		if s1.To.DiskId != s3.To.DiskId {
			t.Fatal("didn't refill random delete")
		}
	} else {
		t.Fatal("Wrong payload type")
	}

}

func TestStateHandlerAsRemote(t *testing.T) {
	outPayload := make(chan model.Payload2, 1)

	myNodeId := test.NonMainNode1
	remoteNode1Id := test.MainNodeId
	remoteNode2Id := test.NonMainNode2

	mapper := model.NewNodeConnectionMapper()
	mapper.SetAll(0, "remoteNode1Address", remoteNode1Id)
	mapper.SetAll(1, "remoteNode2Address", remoteNode2Id)

	stateHandler := datalayer.StateHandler{
		OutPayload:  outPayload,
		MyNodeId:    myNodeId,
		NodeConnMap: mapper,
	}
	stateHandler.Start()

	stateHandler.SetDiskSpace(model.NodeDisk{DiskId: "disk1Id", NodeId: myNodeId}, 2)
	<-outPayload
	stateHandler.SetDiskSpace(model.NodeDisk{DiskId: "disk2Id", NodeId: remoteNode1Id}, 1)
	<-outPayload
	stateHandler.SetDiskSpace(model.NodeDisk{DiskId: "disk3Id", NodeId: remoteNode2Id}, 3)
	<-outPayload

	stateHandler.Saved("block1Id", model.NodeDisk{DiskId: "disk1Id", NodeId: myNodeId})
	sendPayloadMsg := <-outPayload
	if saveParams, ok := sendPayloadMsg.(*datalayer.SavedParams); ok {
		if saveParams.B != "block1Id" {
			t.Error("Invalid BlockId")
		}
		if saveParams.D.DiskId != "disk1Id" || saveParams.D.NodeId != myNodeId {
			t.Error("wrong dest")
		}
	} else {
		t.Error("wrong type")
	}

	stateHandler.Deleted("block1Id", model.NodeDisk{DiskId: "disk1Id", NodeId: myNodeId})
	<-outPayload
}
