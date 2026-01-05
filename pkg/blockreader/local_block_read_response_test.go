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

package blockreader

import (
	"context"
	"tealfs/pkg/model"
	"testing"

	"github.com/google/uuid"
)

func TestLocalBlockReadResponse(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	inReadResults := make(chan (<-chan model.ReadResult))
	localReadResponses := make(chan GetFromDiskResp)
	sends := make(chan model.SendPayloadMsg)
	nodeConnMap := model.NewNodeConnectionMapper()
	localNodeId := model.NewNodeId()
	remoteNodeId := model.NewNodeId()
	blockId1 := model.NewBlockId()
	blockId2 := model.NewBlockId()
	reqId1 := model.GetBlockId(uuid.NewString())
	reqId2 := model.GetBlockId(uuid.NewString())
	lbrr := LocalBlockReadResponses{
		InReadResults:      inReadResults,
		LocalReadResponses: localReadResponses,
		Sends:              sends,
		NodeConnMap:        nodeConnMap,
		NodeId:             localNodeId,
	}
	go lbrr.Start(ctx)

	readResults1 := make(chan model.ReadResult)
	readResults2 := make(chan model.ReadResult)
	inReadResults <- readResults1
	inReadResults <- readResults2

	// disk1.InReads <- model.ReadRequest{
	// 	Caller: localNodeId,
	// 	Ptrs: []model.DiskPointer{{
	// 		NodeId:   localNodeId,
	// 		Disk:     disk1.Id(),
	// 		FileName: string(blockId1),
	// 	}},
	// 	BlockId: blockId1,
	// 	ReqId:   reqId1,
	// }
	ptr1 := model.DiskPointer{
		NodeId:   localNodeId,
		Disk:     "disk1Id",
		FileName: string(blockId1),
	}
	ptrs1 := []model.DiskPointer{ptr1}
	data1 := model.RawData{Ptr: ptr1}
	readResults1 <- model.NewReadResultOk(localNodeId, ptrs1, data1, reqId1, blockId1)

	resp1 := <-localReadResponses
	if resp1.Resp.Id != reqId1 {
		t.Error("invalid request id")
		return
	}

	nodeConnMap.SetAll(model.ConnId(1), "someAddress:123", remoteNodeId)

	ptr2 := model.DiskPointer{
		NodeId:   localNodeId,
		Disk:     "disk1Id",
		FileName: string(blockId1),
	}
	ptrs2 := []model.DiskPointer{ptr2}
	data2 := model.RawData{Ptr: ptr2}
	readResults2 <- model.NewReadResultOk(localNodeId, ptrs2, data2, reqId2, blockId2)


	resp2 := <-sends
	payload := resp2.Payload
	if gfdr, ok := payload.(*GetFromDiskResp); !ok {
		t.Error("sent wrong type")
		return
	} else if gfdr.Resp.Id != reqId2 {
		t.Error("sent wrong message")
		return
	}
}
