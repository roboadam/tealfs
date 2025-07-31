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

package disk

import (
	"context"
	"tealfs/pkg/model"
	"testing"
)

func TestRemoteDiskAdder(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	inAddDiskReq := make(chan model.AddDiskReq)
	outSends := make(chan model.MgrConnsSend)
	nodeConnMap := model.NewNodeConnectionMapper()
	adder := RemoteDiskAdder{
		InAddDiskReq: inAddDiskReq,
		OutSends:     outSends,
		NodeConnMap:  nodeConnMap,
	}
	go adder.Start(ctx)

	req := model.AddDiskReq{
		DiskId: "diskId1",
		Path:   "path1",
		NodeId: "remoteNodeId",
	}

	inAddDiskReq <- req

	select{
	case <-outSends:
		t.Error("should not send if no conns")
	default:
	}

	nodeConnMap.SetAll(0, "address1", "remoteNodeId")
	inAddDiskReq <- req

	sent := <- outSends
	if _, ok := sent.Payload.(*model.AddDiskReq); !ok {
		t.Error("wrong payload type")
	}
}
