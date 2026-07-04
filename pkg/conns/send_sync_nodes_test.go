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
	outPayloads := make(chan model.Payload2)
	mapper := model.NewNodeConnectionMapper()

	sendSyncNodes := SendSyncNodes{
		InSendSyncNodes: inSendSyncNodes,
		OutPayload:      outPayloads,
		NodeConnMapper:  mapper,
		NodeId:          model.NewNodeId(),
	}
	go sendSyncNodes.Start(ctx)

	mapper.SetAll(0, "remoteAddress1", "remoteNodeId1")
	mapper.SetAll(1, "remoteAddress2", "remoteNodeId2")

	inSendSyncNodes <- struct{}{}
	outPayload := <-outPayloads
	if syncNodes, ok := outPayload.(*model.SyncNodes); ok {
		if syncNodes.Destination() != "remoteNodeId1" && syncNodes.Destination() != "remoteNodeId2" {
			t.Fatal("wrong dest")
		}
		if syncNodes.Nodes.Len() != 2 {
			t.Fatal("Expected 2 nodes")
		}
	} else {
		t.Fatal("wrong type")
	}

	outPayload = <-outPayloads
	if syncNodes, ok := outPayload.(*model.SyncNodes); ok {
		if syncNodes.Destination() != "remoteNodeId1" && syncNodes.Destination() != "remoteNodeId2" {
			t.Fatal("wrong dest")
		}
		if syncNodes.Nodes.Len() != 2 {
			t.Fatal("Expected 2 nodes")
		}
	} else {
		t.Fatal("wrong type")
	}
}
