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

package conns

import (
	"context"
	"tealfs/pkg/model"
	"testing"
)

func TestSendSyncNodes(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	inSendSyncNodes := make(chan struct{})
	outSendPayloads := make(chan model.SendPayloadMsg)
	mapper := model.NewNodeConnectionMapper()

	sendSyncNodes := SendSyncNodes{
		InSendSyncNodes: inSendSyncNodes,
		OutSendPayloads: outSendPayloads,
		NodeConnMapper:  mapper,
	}
	go sendSyncNodes.Start(ctx)

	mapper.SetAll(0, "remoteAddress1", "remoteNodeId1")
	mapper.SetAll(1, "remoteAddress2", "remoteNodeId2")

	inSendSyncNodes <- struct{}{}
	sendPayload := <-outSendPayloads
	if sendPayload.ConnId != 0 && sendPayload.ConnId != 1 {
		t.Error("Expected ConnId to be 0")
		return
	}
	switch p := sendPayload.Payload.(type) {
	case *model.SyncNodes:
		if p.Nodes.Len() != 2 {
			t.Error("Expected 2 nodes")
			return
		}
	default:
		t.Error("Unexpected payload", p)
		return
	}

	sendPayload = <-outSendPayloads
	if sendPayload.ConnId != 0 && sendPayload.ConnId != 1 {
		t.Error("Expected ConnId to be 1")
		return
	}
	switch p := sendPayload.Payload.(type) {
	case *model.SyncNodes:
		if p.Nodes.Len() != 2 {
			t.Error("Expected 2 nodes")
			return
		}
	default:
		t.Error("Unexpected payload", p)
		return
	}
}
