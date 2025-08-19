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
	"sync"
	"tealfs/pkg/disk"
	"tealfs/pkg/model"
	"tealfs/pkg/set"
)

type OnDiskBlockIds struct {
	InFetchIds   <-chan AllBlockIdReq
	OutIdResults chan<- AllBlockIdResp

	Disks *set.Set[disk.Disk]
}

func (o *OnDiskBlockIds) Start(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			return
		case req := <-o.InFetchIds:
			o.collectResults(req)
		}
	}
}

func (o *OnDiskBlockIds) collectResults(req AllBlockIdReq) {
	allIds := set.NewSet[model.BlockId]()
	wg := sync.WaitGroup{}
	for _, disk := range o.Disks.GetValues() {
		wg.Add(1)
		go o.readListFromDisk(&disk, &allIds, &wg)
	}
	wg.Wait()
	o.OutIdResults <- AllBlockIdResp{
		Caller:   req.Caller,
		BlockIds: allIds,
		Id:       req.Id,
	}
}

func (o *OnDiskBlockIds) readListFromDisk(d *disk.Disk, allIds *set.Set[model.BlockId], wg *sync.WaitGroup) {
	defer wg.Done()
	d.InListIds <- struct{}{}
	ids := <-d.OutListIds
	allIds.AddAll(&ids)
}
