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

type StoreItReqHandler struct {
	InStoreItReq   <-chan StoreItReq
	OutStoreItResp chan<- StoreItResp

	LocalDisks *set.Set[disk.Disk]
}

func (s *StoreItReqHandler) Start(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			return
		case req := <-s.InStoreItReq:
			s.handleStoreItReq(req)
		}
	}
}

func (s *StoreItReqHandler) handleStoreItReq(req StoreItReq) {
	for _, d := range s.LocalDisks.GetValues() {
		if d.Id() == req.DiskId {
			data, ok := d.Get(req.Cmd.DestBlockId)
			s.sendStoreItResp(req, data, ok)
		}
	}
}

func (s *StoreItReqHandler) sendStoreItResp(req StoreItReq, data []byte, ok bool) {
	s.OutStoreItResp <- StoreItResp{
		Req: req,
		Block: model.Block{
			Id:   req.Cmd.DestBlockId,
			Data: data,
		},
		Ok: ok,
	}
}
