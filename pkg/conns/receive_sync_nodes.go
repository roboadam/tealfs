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

package conns

import (
	"context"
	"tealfs/pkg/model"
)

type ReceiveSyncNodes struct {
	InSyncNodes  <-chan model.SyncNodes
	OutConnectTo chan<- model.ConnectToNodeReq

	NodeConnMapper *model.NodeConnectionMapper
	NodeId         model.NodeId
}

func (r *ReceiveSyncNodes) Start(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			return
		case syncNodes := <-r.InSyncNodes:
			r.handleSyncNodes(&syncNodes)
		}
	}
}

func (r *ReceiveSyncNodes) handleSyncNodes(syncNodes *model.SyncNodes) {
	remoteNodes := syncNodes.GetNodes()
	localNodes := r.NodeConnMapper.Nodes()
	localNodes.Add(r.NodeId)
	missing := remoteNodes.Minus(&localNodes)
	for _, n := range missing.GetValues() {
		address := syncNodes.AddressForNode(n)
		mct := model.ConnectToNodeReq{Address: address}
		r.OutConnectTo <- mct
	}
}
