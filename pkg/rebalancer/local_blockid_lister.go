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

	log "github.com/sirupsen/logrus"
)

type LocalBlockIdLister struct {
	InFetchIds         <-chan AllBlockIdReq
	OutIdLocalResults  chan<- AllBlockIdResp
	OutIdRemoteResults chan<- model.MgrConnsSend

	Disks  *set.Set[disk.Disk]
	NodeId model.NodeId
	Mapper *model.NodeConnectionMapper
}

func (o *LocalBlockIdLister) Start(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			return
		case req := <-o.InFetchIds:
			o.collectResults(req)
		}
	}
}

func (o *LocalBlockIdLister) collectResults(req AllBlockIdReq) {
	allIds := set.NewSet[model.BlockId]()
	wg := sync.WaitGroup{}
	for _, disk := range o.Disks.GetValues() {
		wg.Add(1)
		go o.readListFromDisk(&disk, &allIds, &wg)
	}
	wg.Wait()
	o.sendResults(&AllBlockIdResp{
		Caller:   req.Caller,
		BlockIds: allIds,
		Id:       req.Id,
	})
}

func (o *LocalBlockIdLister) sendResults(resp *AllBlockIdResp) {
	if resp.Caller == o.NodeId {
		o.OutIdLocalResults <- *resp
	} else {
		connId, ok := o.Mapper.ConnForNode(resp.Caller)
		if ok {
			o.OutIdRemoteResults <- model.MgrConnsSend{
				Payload: resp,
				ConnId:  connId,
			}
		} else {
			log.Warn("could not find connection for node")
		}
	}
}

func (o *LocalBlockIdLister) readListFromDisk(d *disk.Disk, allIds *set.Set[model.BlockId], wg *sync.WaitGroup) {
	defer wg.Done()
	d.InListIds <- struct{}{}
	ids := <-d.OutListIds
	allIds.AddAll(&ids)
}
