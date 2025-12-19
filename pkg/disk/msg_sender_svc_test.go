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

package disk_test

import (
	"context"
	"tealfs/pkg/disk"
	"tealfs/pkg/model"
	"testing"
)

func TestExistsSender(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	inAddDiskMsg := make(chan model.AddDiskMsg)
	outAddDiskMsg := make(chan model.AddDiskMsg)
	outRemote := make(chan model.SendPayloadMsg)

	sender := disk.MsgSenderSvc{
		OutRemote:     outRemote,
		NodeId:        "localNodeId",
		NodeConnMap:   model.NewNodeConnectionMapper(),
		InAddDiskMsg:  inAddDiskMsg,
		OutAddDiskMsg: outAddDiskMsg,
	}
	go sender.Start(ctx)

	sender.NodeConnMap.SetAll(0, "someAddress1:123", "remoteNodeId")

	localMsg := model.AddDiskMsg{NodeId: "localNodeId"}

	inAddDiskMsg <- localMsg
	<-outAddDiskMsg

	remoteMsg := model.AddDiskMsg{NodeId: "remoteNodeId"}

	inAddDiskMsg <- remoteMsg
	mcs := <-outRemote

	if mcs.Payload.Type() != model.AddDiskMsgType {
		t.Error("invalid payload type")
	}

	if mcs.Payload.(*model.AddDiskMsg).NodeId != "remoteNodeId" {
		t.Error("invalid node id")
	}
}
