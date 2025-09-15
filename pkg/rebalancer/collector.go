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
	InDiskBlockIds       <-chan OnDiskBlockIdList
	InFilesystemBlockIds <-chan FilesystemBlockIdList
	OutFetchActiveIds    chan<- BalanceReqId
	OutRunCleanup        chan<- BalanceReqId

	onDiskIdsCounter set.Map[BalanceReqId, int]
	OnDiskIds        set.Map[BalanceReqId, OnDiskBlockIdList]
	OnFilesystemIds  set.Map[BalanceReqId, FilesystemBlockIdList]
	Mapper           *model.NodeConnectionMapper
	NodeId           model.NodeId
}

func (c *Collector) Start(ctx context.Context) {
	c.OnDiskIds = set.NewMap[BalanceReqId, OnDiskBlockIdList]()
	c.OnFilesystemIds = set.NewMap[BalanceReqId, FilesystemBlockIdList]()
	c.onDiskIdsCounter = set.NewMap[BalanceReqId, int]()

	for {
		select {
		case <-ctx.Done():
			return
		case resp := <-c.InDiskBlockIds:
			all, ok := c.OnDiskIds.Get(resp.BalanceReqId)
			if !ok {
				all = OnDiskBlockIdList{
					Caller:       resp.Caller,
					BlockIds:     set.NewSet[model.BlockId](),
					BalanceReqId: resp.BalanceReqId,
				}
			}
			all.BlockIds.AddAll(&resp.BlockIds)
			c.OnDiskIds.Add(resp.BalanceReqId, all)
			count := c.increment(resp.BalanceReqId)
			if count >= c.expectedOnDiskMsgs() {
				c.OutFetchActiveIds <- resp.BalanceReqId
			}
		case resp := <-c.InFilesystemBlockIds:
			c.OnFilesystemIds.Add(resp.BalanceReqId, resp)
			c.OutRunCleanup <- resp.BalanceReqId
		}
	}
}

func (c *Collector) expectedOnDiskMsgs() int {
	conns := c.Mapper.Connections()
	return 1 + conns.Len()
}

func (c *Collector) increment(id BalanceReqId) int {
	count, ok := c.onDiskIdsCounter.Get(id)
	if !ok {
		count = 0
	}
	count++
	c.onDiskIdsCounter.Add(id, count)
	return count
}
