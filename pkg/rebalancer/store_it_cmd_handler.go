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
	"tealfs/pkg/disk"
	"tealfs/pkg/model"
	"tealfs/pkg/set"
)

type StoreItCmdHandler struct {
	InStoreItCmd  <-chan StoreItCmd
	InStoreItResp <-chan StoreItResp
	OutStoreItReq chan<- StoreItReq

	AllDiskIds  *set.Set[model.AddDiskReq]
	LocalDisks  *set.Set[disk.Disk]
	bookKeeping map[BalanceReqId]map[model.BlockId]*set.Set[model.AddDiskReq]
}

func (s *StoreItCmdHandler) Start(ctx context.Context) {
	s.bookKeeping = make(map[BalanceReqId]map[model.BlockId]*set.Set[model.AddDiskReq])
	for {
		select {
		case <-ctx.Done():
			return
		case cmd := <-s.InStoreItCmd:
			s.handleStoreItCmd(cmd)
		case resp := <-s.InStoreItResp:
			s.handleStoreItResp(resp)
		}
	}
}

func (s *StoreItCmdHandler) handleStoreItResp(resp StoreItResp) {
	if resp.Ok {
		for _, d := range s.LocalDisks.GetValues() {
			if d.Id() == resp.Req.DestDiskId {
			}
		}
	} else {
		s.sendNextStoreItReq(resp.Req)
	}

}

func (s *StoreItCmdHandler) handleStoreItCmd(cmd StoreItCmd) {
	s.initBookKeeping(cmd)
	s.bookKeeping[cmd.BalanceReqId][cmd.DestBlockId] = s.AllDiskIds
	s.sendNextStoreItReq(cmd)
}

func (s *StoreItCmdHandler) sendNextStoreItReq(cmd StoreItCmd) {
	dests := s.bookKeeping[cmd.BalanceReqId][cmd.DestBlockId]
	for {
		next, remainder, ok := dests.Pop()
		s.bookKeeping[cmd.BalanceReqId][cmd.DestBlockId] = remainder

		if !ok {
			break
		}
		if next.NodeId == cmd.Caller {
			continue
		}

		s.OutStoreItReq <- StoreItReq{
			Req:    cmd,
			NodeId: next.NodeId,
			DiskId: next.DiskId,
		}
		break
	}
}

func (s *StoreItCmdHandler) initBookKeeping(cmd StoreItCmd) {
	_, ok := s.bookKeeping[cmd.BalanceReqId]
	if !ok {
		s.bookKeeping[cmd.BalanceReqId] = make(map[model.BlockId]*set.Set[model.AddDiskReq])
	}
}
