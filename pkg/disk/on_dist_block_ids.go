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

package disk

import (
	"context"
	"tealfs/pkg/model"
	"tealfs/pkg/rebalancer"
	"tealfs/pkg/set"
)

type OnDiskBlockIds struct {
	InFetchIds   <-chan rebalancer.AllBlockIdReq
	OutIdResults chan<- rebalancer.AllBlockIdResp

	outListIds   set.Set[chan<- rebalancer.AllBlockIdReq]
	inIdResult   set.Set[<-chan rebalancer.AllBlockIdResp]

	resultsHolder map[rebalancer.AllBlockId]rebalancer.AllBlockIdResp
}

func (o *OnDiskBlockIds) Start(ctx context.Context) {
	o.resultsHolder = make(map[rebalancer.AllBlockId]rebalancer.AllBlockIdResp)
	for {
		select {
		case <-ctx.Done():
			return
		case req := <-o.InFetchIds:
			o.forwardToDisks(req)
		}
	}
}

func (o *OnDiskBlockIds) forwardToDisks(req rebalancer.AllBlockIdReq) {
	o.resultsHolder[req.Id] = rebalancer.AllBlockIdResp{
		Caller:   req.Caller,
		BlockIds: set.NewSet[model.BlockId](),
		Id:       req.Id,
	}
	for _, out := range o.OutListIds.GetValues() {
		out <- req
	}
}
