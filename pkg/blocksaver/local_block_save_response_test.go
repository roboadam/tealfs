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

package blocksaver

import (
	"context"
	"tealfs/pkg/model"
	"testing"

	"github.com/google/uuid"
)

func TestLocalBlockSaveResponse(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	nodeId := model.NewNodeId()
	remoteNodeId := model.NewNodeId()

	resp := make(chan SaveToDiskResp)
	sends := make(chan model.SendPayloadMsg)
	inWriteResults := make(chan (<-chan model.WriteResult))

	lbsr := LocalBlockSaveResponses{
		InWriteResults:      inWriteResults,
		LocalWriteResponses: resp,
		Sends:               sends,
		NodeConnMap:         model.NewNodeConnectionMapper(),
		NodeId:              nodeId,
	}

	go lbsr.Start(ctx)

	writeResults1 := make(chan model.WriteResult)
	writeResults2 := make(chan model.WriteResult)
	inWriteResults <- writeResults1
	inWriteResults <- writeResults2

	lbsr.NodeConnMap.SetAll(1, "some-address:123", remoteNodeId)

	putBlockId := model.PutBlockId(uuid.NewString())

	ptr := model.DiskPointer{
		NodeId:   nodeId,
		Disk:     "diskId1",
		FileName: uuid.NewString(),
	}
	writeResults1 <- model.NewWriteResultOk(ptr, nodeId, putBlockId)

	wr := <-resp
	if wr.Resp.Id != putBlockId {
		t.Error("Unknown put block id")
		return
	}

	putBlockId2 := model.PutBlockId(uuid.NewString())
	writeResults2 <- model.NewWriteResultErr(
		"some error happened",
		nodeId,
		putBlockId2,
	)

	wr = <-resp
	if wr.Resp.Id != putBlockId2 {
		t.Error("Unknown put block id")
		return
	}

	putBlockId3 := model.PutBlockId(uuid.NewString())
	writeResults1 <- model.NewWriteResultOk(
		model.DiskPointer{
			NodeId:   nodeId,
			Disk:     "diskId1",
			FileName: uuid.NewString(),
		},
		remoteNodeId,
		putBlockId3,
	)

	payload := <-sends
	if payload.Payload.Type() != model.SaveToDiskResp {
		t.Error("unknown send payload")
		return
	}
}
