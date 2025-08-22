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

type Collector struct {
	InDiskBlockIds          <-chan AllBlockIdResp
	InFilesystemBlockIds    <-chan AllBlockIdResp
	OutLocalAllBlockIdResp  chan<- AllBlockIdResp
	OutRemoteAllBlockIdResp chan<- AllBlockIdResp
	OutFetchActiveIds       chan<- struct{}

	OnDiskIds        set.Map[AllBlockId, AllBlockIdResp]
	onDiskIdsCounter set.Map[AllBlockId, int]
	OnFilesystemIds  set.Map[AllBlockId, AllBlockIdResp]
	Mapper           *model.NodeConnectionMapper
	NodeId           model.NodeId
}

func (c *Collector) Start(ctx context.Context) {
	c.OnDiskIds = set.NewMap[AllBlockId, AllBlockIdResp]()
	c.onDiskIdsCounter = set.NewMap[AllBlockId, int]()

	for {
		select {
		case <-ctx.Done():
			return
		case resp := <-c.InDiskBlockIds:
			all, ok := c.OnDiskIds.Get(resp.Id)
			if !ok {
				all = AllBlockIdResp{
					Caller:   resp.Caller,
					BlockIds: set.NewSet[model.BlockId](),
					Id:       resp.Id,
				}
			}
			all.BlockIds.AddAll(&resp.BlockIds)
			c.OnDiskIds.Add(resp.Id, all)
			c.onDiskIdsCounter.Get(resp.Id)
			c.onDiskIdsCounter.Add(resp.Id, 1)

			if c.
			// c.collector[resp.Id] = append(c.collector[resp.Id], resp)
			// conns := c.Mapper.Connections()
			// finalCount := 1 + conns.Len()
			// if len(c.collector[resp.Id]) == finalCount {
			// 	resps := c.collector[resp.Id]
			// 	delete(c.collector, resp.Id)
			// 	c.sendResp(aggregate(resps))
			// }
		}
	}
}

func (c *Collector) sendResp(resp *AllBlockIdResp) {
	if c.NodeId == resp.Caller {
		c.OutLocalAllBlockIdResp <- *resp
	} else {
		c.OutRemoteAllBlockIdResp <- *resp
	}
}

func aggregate(resps []AllBlockIdResp) *AllBlockIdResp {
	if len(resps) == 0 {
		return &AllBlockIdResp{}
	}
	result := AllBlockIdResp{
		BlockIds: set.NewSet[model.BlockId](),
		Caller:   resps[0].Caller,
		Id:       resps[0].Id,
	}
	for _, resp := range resps {
		result.BlockIds.AddAll(&resp.BlockIds)
	}
	return &result
}
