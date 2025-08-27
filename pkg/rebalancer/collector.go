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
	InDiskBlockIds       <-chan AllBlockIdResp
	InFilesystemBlockIds <-chan AllBlockIdResp
	OutFetchActiveIds    chan<- AllBlockId
	OutRunCleanup        chan<- AllBlockId

	onDiskIdsCounter set.Map[AllBlockId, int]
	OnDiskIds        set.Map[AllBlockId, AllBlockIdResp]
	OnFilesystemIds  set.Map[AllBlockId, AllBlockIdResp]
	Mapper           *model.NodeConnectionMapper
	NodeId           model.NodeId
}

func (c *Collector) Start(ctx context.Context) {
	c.OnDiskIds = set.NewMap[AllBlockId, AllBlockIdResp]()
	c.OnFilesystemIds = set.NewMap[AllBlockId, AllBlockIdResp]()
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
			count := c.increment(resp.Id)
			if count >= c.expectedOnDiskMsgs() {
				c.OutFetchActiveIds <- resp.Id
			}
		case resp := <-c.InFilesystemBlockIds:
			c.OnFilesystemIds.Add(resp.Id, resp)
			c.OutRunCleanup <- resp.Id
		}
	}
}

func (c *Collector) expectedOnDiskMsgs() int {
	conns := c.Mapper.Connections()
	return 1 + conns.Len()
}

func (c *Collector) increment(id AllBlockId) int {
	count, ok := c.onDiskIdsCounter.Get(id)
	if !ok {
		count = 0
	}
	count++
	c.onDiskIdsCounter.Add(id, count)
	return count
}
