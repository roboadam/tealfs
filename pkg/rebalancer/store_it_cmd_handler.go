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

	log "github.com/sirupsen/logrus"
)

type StoreItCmdHandler struct {
	InStoreItCmd  <-chan StoreItCmd
	InStoreItResp <-chan StoreItResp

	OutStoreItReq chan<- StoreItReq
	OutExistsResp chan<- ExistsResp

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
			if d.Id() == resp.Req.DiskId {
				ok := d.Save(resp.Block.Data, resp.Block.Id)
				if ok {
					s.OutExistsResp <- ExistsResp{
						Req: resp.Req.Cmd.ExistsReq,
						Ok:  true,
					}
				} else {
					log.Warn("Failed to save block")
				}
			}
		}
	} else {
		s.sendNextStoreItReq(resp.Req.Cmd)
	}

}

func (s *StoreItCmdHandler) handleStoreItCmd(cmd StoreItCmd) {
	s.initBookKeeping(cmd)
	s.bookKeeping[cmd.BalanceReqId][cmd.DestBlockId] = s.AllDiskIds
	s.sendNextStoreItReq(cmd)
}

func (s *StoreItCmdHandler) sendNextStoreItReq(cmd StoreItCmd) {
	remainder := s.bookKeeping[cmd.BalanceReqId][cmd.DestBlockId]
	for {
		next, remainder, ok := remainder.Pop()
		s.bookKeeping[cmd.BalanceReqId][cmd.DestBlockId] = remainder

		// If there are no more possible locations the block could be so give up
		if !ok {
			break
		}

		// You don't need to bother checking the intended destination for the data
		if next.DiskId == cmd.DestDiskId {
			continue
		}

		s.OutStoreItReq <- StoreItReq{
			Cmd:    cmd,
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
