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

func TestReceiveSyncNotes(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	inSyncNodes := make(chan model.SyncNodes)
	outConnectTo := make(chan model.ConnectToNodeReq)
	nodeConnMapper := model.NewNodeConnectionMapper()
	nodeConnMapper.SetAll(0, "remoteAddress1", "remoteNodeId1")

	receiveSyncNodes := ReceiveSyncNodes{
		InSyncNodes:    inSyncNodes,
		OutConnectTo:   outConnectTo,
		NodeConnMapper: nodeConnMapper,
		NodeId:         "nodeId",
	}
	go receiveSyncNodes.Start(ctx)

	syncNodes := model.NewSyncNodes()
	syncNodes.Nodes.Add(struct {
		Node    model.NodeId
		Address string
	}{Node: "remoteNodeId1", Address: "remoteAddress1"})
	syncNodes.Nodes.Add(struct {
		Node    model.NodeId
		Address string
	}{Node: "remoteNodeId2", Address: "remoteAddress2"})

	inSyncNodes <- syncNodes

	connToReq := <-outConnectTo
	if connToReq.Address != "remoteAddress2" {
		t.Error("Expected remoteAddress2")
		return
	}

	select {
	case <-outConnectTo:
		t.Error("should only connect to the new address")
	default:
	}
}
