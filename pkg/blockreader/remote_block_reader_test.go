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

func TestRemoteBlockReader(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	req := make(chan GetFromDiskReq)
	sends := make(chan model.SendPayloadMsg)
	noConnResp := make(chan GetFromDiskResp)
	nodeId := model.NewNodeId()
	diskId := model.DiskId(uuid.NewString())
	reqId := model.GetBlockId(uuid.NewString())
	blockId := model.NewBlockId()

	br := RemoteBlockReader{
		Req:         req,
		Sends:       sends,
		NoConnResp:  noConnResp,
		NodeConnMap: model.NewNodeConnectionMapper(),
	}

	go br.Start(ctx)

	gfdr := GetFromDiskReq{
		Caller: nodeId,
		Dest: Dest{
			NodeId: nodeId,
			DiskId: diskId,
		},
		Req: model.GetBlockReq{
			Id:      reqId,
			BlockId: blockId,
		},
	}
	req <- gfdr

	resp := <-noConnResp
	if resp.Resp.Err == nil || resp.Resp.Id != reqId {
		t.Error("should be an error")
		return
	}

	br.NodeConnMap.SetAll(0, "someAddress:123", nodeId)

	req <- gfdr
	sent := <-sends
	payload := sent.Payload

	if g, ok := payload.(*GetFromDiskReq); ok {
		if g.Req.Id != gfdr.Req.Id {
			t.Error("invalid send")
			return
		}
	} else {
		t.Error("invalid type")
		return
	}
}
