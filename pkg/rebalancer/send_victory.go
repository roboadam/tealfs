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

package rebalancer

import (
	"context"
	"tealfs/pkg/model"
)

type SendVictory struct {
	NodeId model.NodeId
	Mapper *model.NodeConnectionMapper

	InTrigger <-chan struct{}
	OutSends  chan<- model.MgrConnsSend
}

func (s *SendVictory) Start(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			return
		case <-s.InTrigger:
			s.sendVictory()
		}
	}
}

func (s *SendVictory) sendVictory() {
	conns := s.Mapper.Connections()
	for _, conn := range conns.GetValues() {
		s.OutSends <- model.MgrConnsSend{
			ConnId:  conn,
			Payload: &Victory{NodeID: s.NodeId},
		}
	}
}
