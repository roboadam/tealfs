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

	"github.com/google/uuid"
)

type Startup struct {
	InTrigger   <-chan struct{}
	OutSends    chan<- model.MgrConnsSend
	OutLocalReq chan<- AllBlockIdReq

	NodeId model.NodeId
	Mapper *model.NodeConnectionMapper
	Disks  *set.Set[model.AddDiskReq]
}

func (s *Startup) Start(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			return
		case <-s.InTrigger:
			if s.IAmPrimary() && s.Mapper.AreAllConnected() {
				req := AllBlockIdReq{
					Caller: s.NodeId,
					Id:     AllBlockId(uuid.NewString()),
				}
				conns := s.Mapper.Connections()
				for _, conn := range conns.GetValues() {
					s.OutSends <- model.MgrConnsSend{
						Payload: &req,
						ConnId:  conn,
					}
				}
			}
		}
	}
}

func (s *Startup) IAmPrimary() bool {
	return PrimaryNode(s.Disks) == s.NodeId
}
