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
	"tealfs/pkg/datalayer"
	"tealfs/pkg/disk"
	"tealfs/pkg/model"
	"tealfs/pkg/set"
	"testing"
)

func TestSaveRequestHandler(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	inSaveRequests := make(chan datalayer.SaveRequest)
	inDataforSaveRequest := make(chan datalayer.DataForSaveRequest)
	disks := set.NewSet[disk.Disk]()
	nodeId := model.NodeId("nodeId")
	outDataforSaveRequest := make(chan datalayer.DataForSaveRequest)
	outSends := make(chan model.SendPayloadMsg)
	nodeConnMap := model.NewNodeConnectionMapper()
	stateHandler := MockSaver{}
	diskIdFrom := model.DiskId("disk1")
	diskSvc := disk.New(disk.NewPath("foo", &disk.MockFileOps{}), nodeId, diskIdFrom, ctx)
	disks.Add(diskSvc)
	diskIdTo := model.DiskId("disk2")

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
		From:    []datalayer.Dest{},
		BlockId: "",
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
