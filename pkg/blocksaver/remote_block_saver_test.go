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

func TestRemoteBlockSaver(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	req := make(chan SaveToDiskReq)
	sends := make(chan model.SendPayloadMsg)
	noConnResp := make(chan SaveToDiskResp)
	nodeConnMap := model.NewNodeConnectionMapper()

	rbs := RemoteBlockSaver{
		Req:         req,
		Sends:       sends,
		NoConnResp:  noConnResp,
		NodeConnMap: nodeConnMap,
	}

	go rbs.Start(ctx)

	localNodeId := model.NewNodeId()
	connId := model.ConnId(4)
	remoteNodeId := model.NewNodeId()
	address := "some-address:1234"

	connId2 := model.ConnId(5)
	remoteNodeId2 := model.NewNodeId()
	address2 := "other-address:1234"

	nodeConnMap.SetAll(connId, address, remoteNodeId)
	nodeConnMap.SetAll(connId2, address2, remoteNodeId2)

	req <- SaveToDiskReq{
		Caller: localNodeId,
		Dest: Dest{
			NodeId: remoteNodeId,
			DiskId: model.DiskId(uuid.NewString()),
		},
		Req: model.NewPutBlockReq(model.Block{
			Id:   model.NewBlockId(),
			Data: []byte{1, 2, 3, 4, 5},
		}),
	}

	mcs := <-sends

	if mcs.ConnId != connId {
		t.Error("Invalid send")
		return
	}
	if _, ok := mcs.Payload.(*SaveToDiskReq); !ok {
		t.Error("Invalid send")
		return
	}

	req <- SaveToDiskReq{
		Caller: localNodeId,
		Dest: Dest{
			NodeId: model.NewNodeId(),
			DiskId: model.DiskId(uuid.NewString()),
		},
		Req: model.NewPutBlockReq(model.Block{
			Id:   model.NewBlockId(),
			Data: []byte{1, 2, 3, 4, 5},
		}),
	}

	noConn := <-noConnResp

	if noConn.Resp.Err == nil {
		t.Error("Expected an error")
		return
	}

}
