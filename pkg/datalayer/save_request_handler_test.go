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
	"bytes"
	"context"
	"fmt"
	"tealfs/pkg/datalayer"
	"tealfs/pkg/disk"
	"tealfs/pkg/model"
	"tealfs/pkg/set"
	"testing"
)

func TestSaveRequestHandlerSaveRequest(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	fileOps := disk.MockFileOps{}
	inSaveRequests := make(chan datalayer.SaveRequest)
	inDataforSaveRequest := make(chan datalayer.DataForSaveRequest)
	disks := set.NewSet[disk.Disk]()
	nodeIdLocal := model.NodeId("nodeId")
	outDataforSaveRequest := make(chan datalayer.DataForSaveRequest)
	outSends := make(chan model.SendPayloadMsg)
	outPayload := make(chan model.Payload2)
	nodeConnMap := model.NewNodeConnectionMapper()
	stateHandler := MockSaver{}
	diskIdFrom := model.DiskId("disk1")
	diskIdTo := model.DiskId("disk2")
	pathFrom := "pathFrom"
	pathTo := "pathTo"
	diskFrom := disk.New(disk.NewPath(pathFrom, &fileOps), nodeIdLocal, diskIdFrom, ctx)
	diskTo := disk.New(disk.NewPath(pathTo, &fileOps), nodeIdLocal, diskIdTo, ctx)
	disks.Add(diskFrom)
	disks.Add(diskTo)
	blockId := model.BlockId("blockId")
	fileData := []byte{1, 2, 3, 4, 5}

	fileOps.WriteFile(fmt.Sprint(pathFrom, "/", blockId), fileData)

	s := datalayer.SaveRequestHandler{
		InSaveRequests:        inSaveRequests,
		InDataForSaveRequest:  inDataforSaveRequest,
		Disks:                 &disks,
		NodeId:                nodeIdLocal,
		OutDataForSaveRequest: outDataforSaveRequest,
		OutSends:              outSends,
		OutPayload:            outPayload,
		NodeConnMap:           nodeConnMap,
		StateHandler:          &stateHandler,
	}

	go s.Start(ctx)

	inSaveRequests <- datalayer.SaveRequest{
		To:      model.NodeDisk{DiskId: diskIdTo, NodeId: nodeIdLocal},
		From:    []model.NodeDisk{{DiskId: diskIdFrom, NodeId: nodeIdLocal}},
		BlockId: blockId,
	}

	data := <-outPayload
	if req, ok := data.(*datalayer.DataForSaveRequest); ok {
		if !bytes.Equal(fileData, req.Data) {
			t.Error("invalid data")
		}
	} else {
		t.Error("Invalid type")
	}

	diskIdTo = model.DiskId("disk3")
	nodeIdRemote := model.NodeId("nodeIdRemote")
	diskTo = disk.New(disk.NewPath(pathTo, &fileOps), nodeIdRemote, diskIdTo, ctx)
	disks.Add(diskTo)
	nodeConnMap.SetAll(0, "address", nodeIdRemote)

	inSaveRequests <- datalayer.SaveRequest{
		To:      model.NodeDisk{DiskId: diskIdTo, NodeId: nodeIdRemote},
		From:    []model.NodeDisk{{DiskId: diskIdFrom, NodeId: nodeIdLocal}},
		BlockId: blockId,
	}

	payload := <-outPayload
	if req, ok := payload.(*datalayer.DataForSaveRequest); ok {
		if !bytes.Equal(fileData, req.Data) {
			t.Error("invalid data")
		}
	} else {
		t.Error("Invalid type")
	}
}

func TestSaveRequestHandlerDataForSaveRequest(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	fileOps := disk.MockFileOps{}
	inSaveRequests := make(chan datalayer.SaveRequest)
	inDataForSaveRequest := make(chan datalayer.DataForSaveRequest)
	disks := set.NewSet[disk.Disk]()
	nodeIdLocal := model.NodeId("nodeId")
	outDataForSaveRequest := make(chan datalayer.DataForSaveRequest)
	outSends := make(chan model.SendPayloadMsg)
	nodeConnMap := model.NewNodeConnectionMapper()
	savedCh := make(chan struct{}, 1)
	stateHandler := MockSaver{SavedCh: savedCh}
	diskId := model.DiskId("disk1")
	path := "somePath"
	d := disk.New(disk.NewPath(path, &fileOps), nodeIdLocal, diskId, ctx)
	disks.Add(d)
	blockId := model.BlockId("blockId")
	fileData := []byte{1, 2, 3, 4, 5}

	s := datalayer.SaveRequestHandler{
		InSaveRequests:        inSaveRequests,
		InDataForSaveRequest:  inDataForSaveRequest,
		Disks:                 &disks,
		NodeId:                nodeIdLocal,
		OutDataForSaveRequest: outDataForSaveRequest,
		OutSends:              outSends,
		NodeConnMap:           nodeConnMap,
		StateHandler:          &stateHandler,
	}

	go s.Start(ctx)

	inDataForSaveRequest <- datalayer.DataForSaveRequest{
		SaveRequest: datalayer.SaveRequest{
			To:      model.NodeDisk{DiskId: diskId, NodeId: nodeIdLocal},
			BlockId: blockId,
		},
		Data: fileData,
	}

	<-savedCh

	if stateHandler.BlockId != blockId {
		t.Error("invalid block id")
	}
	if stateHandler.Dest != (model.NodeDisk{DiskId: diskId, NodeId: nodeIdLocal}) {
		t.Error("invalid dest")
	}
}

type MockSaver struct {
	BlockId model.BlockId
	Dest    model.NodeDisk
	Count   int
	SavedCh chan struct{}
}

func (m *MockSaver) Saved(blockId model.BlockId, d model.NodeDisk) {
	m.Count++
	m.BlockId = blockId
	m.Dest = d
	if m.SavedCh != nil {
		m.SavedCh <- struct{}{}
	}
}
