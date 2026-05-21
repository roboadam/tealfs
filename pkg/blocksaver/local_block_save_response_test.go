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
	"tealfs/pkg/datalayer"
	"tealfs/pkg/model"
	"tealfs/pkg/test"
	"testing"

	"github.com/google/uuid"
)

func TestLocalBlockSaveResponse(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	nodeId := test.NonMainNode1
	remoteNodeId := test.MainNodeId

	resp := make(chan SaveToDiskResp, 1)
	sends := make(chan model.SendPayloadMsg, 1)
	inWriteResults := make(chan (<-chan model.WriteResult), 1)
	outSaveRequest := make(chan datalayer.SaveRequest, 1)
	outPayload := make(chan model.Payload2)
	nodeConnMapper := model.NewNodeConnectionMapper()

	stateHandler := datalayer.StateHandler{
		OutSaveRequest: outSaveRequest,
		OutPayload:     outPayload,
		MyNodeId:       nodeId,
		NodeConnMap:    nodeConnMapper,
	}

	lbsr := LocalBlockSaveResponses{
		InWriteResults:      inWriteResults,
		LocalWriteResponses: resp,
		Sends:               sends,
		NodeConnMap:         nodeConnMapper,
		NodeId:              nodeId,
		StateHandler:        &stateHandler,
	}

	go lbsr.Start(ctx)

	go func(ctx context.Context) {
		for {
			select {
			case <-ctx.Done():
				return
			case <-outPayload:
			}
		}
	}(ctx)

	writeResults1 := make(chan model.WriteResult, 1)
	writeResults2 := make(chan model.WriteResult, 1)
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
	saveParamsPayload := <-sends
	if saveParams, ok := saveParamsPayload.Payload.(datalayer.SavedParams); ok {
		if saveParams.D.NodeId != nodeId {
			t.Errorf("wrong dest. expected %s, got %s", nodeId, saveParams.D.NodeId)
		}
	} else {
		t.Error("wrong type")
	}

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

	payloadSaveParams := <-sends
	if _, ok := payloadSaveParams.Payload.(datalayer.SavedParams); !ok {
		t.Error("unknown send payload")
		return
	}

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

	payloadSaveParams = <-sends
	if _, ok := payloadSaveParams.Payload.(datalayer.SavedParams); !ok {
		t.Error("unknown send payload")
		return
	}

	payload := <-sends
	if _, ok := payload.Payload.(*SaveToDiskResp); !ok {
		t.Error("unknown send payload")
		return
	}
}
