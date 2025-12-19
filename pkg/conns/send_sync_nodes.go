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

type SendSyncNodes struct {
	InSendSyncNodes <-chan struct{}
	OutSendPayloads chan<- model.SendPayloadMsg

	NodeConnMapper *model.NodeConnectionMapper
}

func (s *SendSyncNodes) Start(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			return
		case <-s.InSendSyncNodes:
			s.send(s.syncNodesPayloadToSend())
		}
	}
}

func (s *SendSyncNodes) send(syncNodes *model.SyncNodes) {
	connections := s.NodeConnMapper.Connections()
	for _, connId := range connections.GetValues() {
		s.OutSendPayloads <- model.SendPayloadMsg{
			ConnId:  connId,
			Payload: syncNodes,
		}

	}
}

func (s *SendSyncNodes) syncNodesPayloadToSend() *model.SyncNodes {
	result := model.NewSyncNodes()
	addressesAndNodes := s.NodeConnMapper.NodesWithAddress()
	for _, an := range addressesAndNodes {
		result.Nodes.Add(struct {
			Node    model.NodeId
			Address string
		}{Node: an.J, Address: an.K})
	}
	return &result
}
