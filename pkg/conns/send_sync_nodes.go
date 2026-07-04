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
	"tealfs/pkg/set"
)

type SendSyncNodes struct {
	InSendSyncNodes <-chan struct{}
	OutPayload      chan<- model.Payload2
	NodeId          model.NodeId

	NodeConnMapper *model.NodeConnectionMapper
}

func (s *SendSyncNodes) Start(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			return
		case <-s.InSendSyncNodes:
			s.send()
		}
	}
}

func (s *SendSyncNodes) send() {
	result := make([]struct {
		Node    model.NodeId
		Address string
	}, 0)
	addressesAndNodes := s.NodeConnMapper.NodesWithAddress()
	for _, an := range addressesAndNodes {
		nodeAddress := struct {
			Node    model.NodeId
			Address string
		}{Node: an.J, Address: an.K}
		result = append(result, nodeAddress)
	}

	nodes := s.NodeConnMapper.ConnectedNodes()
	for _, node := range nodes.GetValues() {
		sn := model.NewSyncNodes(node)
		sn.Nodes = set.NewSetFromSlice(result)
		s.OutPayload <- &sn
	}
}
