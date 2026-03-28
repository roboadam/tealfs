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

func TestSaveRequestHandler(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	fileOps := disk.MockFileOps{}
	inSaveRequests := make(chan datalayer.SaveRequest)
	inDataforSaveRequest := make(chan datalayer.DataForSaveRequest)
	disks := set.NewSet[disk.Disk]()
	nodeId := model.NodeId("nodeId")
	outDataforSaveRequest := make(chan datalayer.DataForSaveRequest)
	outSends := make(chan model.SendPayloadMsg)
	nodeConnMap := model.NewNodeConnectionMapper()
	stateHandler := MockSaver{}
	diskIdFrom := model.DiskId("disk1")
	diskIdTo := model.DiskId("disk2")
	pathFrom := "pathFrom"
	pathTo := "pathTo"
	diskFrom := disk.New(disk.NewPath(pathFrom, &fileOps), nodeId, diskIdFrom, ctx)
	diskTo := disk.New(disk.NewPath(pathTo, &fileOps), nodeId, diskIdTo, ctx)
	disks.Add(diskFrom)
	disks.Add(diskTo)
	blockId := model.BlockId("blockId")
	fileData := []byte{1, 2, 3, 4, 5}

	fileOps.WriteFile(fmt.Sprint(pathFrom, "/", blockId), fileData)

	s := datalayer.SaveRequestHandler{
		InSaveRequests:        inSaveRequests,
		InDataforSaveRequest:  inDataforSaveRequest,
		Disks:                 &disks,
		NodeId:                nodeId,
		OutDataforSaveRequest: outDataforSaveRequest,
		OutSends:              outSends,
		NodeConnMap:           nodeConnMap,
		StateHandler:          &stateHandler,
	}

	go s.Start(ctx)

	inSaveRequests <- datalayer.SaveRequest{
		To:      datalayer.Dest{DiskId: diskIdTo, NodeId: nodeId},
		From:    []datalayer.Dest{{DiskId: diskIdFrom, NodeId: nodeId}},
		BlockId: blockId,
	}

	data := <-outDataforSaveRequest
	if !bytes.Equal(fileData, data.Data) {
		t.Error("invalid data")
	}
}

type MockSaver struct {
	BlockId model.BlockId
	Dest    datalayer.Dest
	Count   int
}

func (m *MockSaver) Saved(blockId model.BlockId, d datalayer.Dest) {
	m.Count++
	m.BlockId = blockId
	m.Dest = d
}
