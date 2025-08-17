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
	"tealfs/pkg/set"
)

type StartElection struct {
	NodeId model.NodeId
	Mapper *model.NodeConnectionMapper

	InNodes  <-chan set.Set[model.NodeId]
	OutSends chan<- model.MgrConnsSend
}

func (s *StartElection) Start(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			return
		case n := <-s.InNodes:
			s.sendElection(n)
		}
	}
}

func (s *StartElection) sendElection(nodes set.Set[model.NodeId]) {
	for _, node := range nodes.GetValues() {
		if connId, ok := s.Mapper.ConnForNode(node); ok {
			s.OutSends <- model.MgrConnsSend{
				ConnId:  connId,
				Payload: &Election{NodeID: s.NodeId},
			}
		}
	}
}
