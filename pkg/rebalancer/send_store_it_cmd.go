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
	"tealfs/pkg/disk/dist"
	"tealfs/pkg/model"
	"tealfs/pkg/set"
)

type SendStoreItCmd struct {
	InStoreItCmd    <-chan StoreItCmd
	OutSends        chan<- model.MgrConnsSend
	OutLocalStoreIt chan<- StoreItCmd

	Distributer *dist.MirrorDistributer
	Conns       *model.NodeConnectionMapper
	NodeId      model.NodeId
}

func (s *SendStoreItCmd) Start(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			return
		case cmd := <-s.InStoreItCmd:
			nodes := s.nodesForCmd(cmd)
			if nodes.Contains(s.NodeId) {
				s.OutLocalStoreIt <- cmd
			} else {
				s.remoteSend(&cmd, nodes)
			}
		}
	}
}

func (s *SendStoreItCmd) remoteSend(cmd *StoreItCmd, nodes *set.Set[model.NodeId]) {
	conns := s.Conns.ConnsForNodes(*nodes)
	for _, conn := range conns.GetValues() {
		s.OutSends <- model.MgrConnsSend{
			Payload: cmd,
			ConnId:  conn,
		}
	}
}

func (s *SendStoreItCmd) nodesForCmd(cmd StoreItCmd) *set.Set[model.NodeId] {
	nodes := set.NewSet[model.NodeId]()
	ptrs := s.Distributer.ReadPointersForId(cmd.BlockId)
	for _, ptr := range ptrs {
		nodes.Add(ptr.NodeId)
	}
	return &nodes
}
