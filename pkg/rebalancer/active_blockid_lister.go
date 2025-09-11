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

type ActiveBlockIdLister struct {
	InFetchIds       <-chan BalanceReq
	OutLocalResults  chan<- BlockIdList
	OutRemoteResults chan<- model.MgrConnsSend

	NodeId model.NodeId
	Mapper *model.NodeConnectionMapper
}

func (l *ActiveBlockIdLister) Start(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			return
		case req := <-l.InFetchIds:
			l.collectResults(req)
		}
	}
}

func (l *ActiveBlockIdLister) collectResults(req BalanceReq) {
	l.sendResults(&BlockIdList{
		Caller:       req.Caller,
		BlockIds:     allIds,
		BalanceReqId: req.BalanceReqId,
	})
}

func (l *ActiveBlockIdLister) sendResults(resp *BlockIdList) {
	if resp.Caller == l.NodeId {
		l.OutLocalResults <- *resp
	} else {
		connId, ok := l.Mapper.ConnForNode(resp.Caller)
		if ok {
			l.OutRemoteResults <- model.MgrConnsSend{
				Payload: resp,
				ConnId:  connId,
			}
		} else {
			log.Warn("could not find connection for node")
		}
	}
}

func (l *ActiveBlockIdLister) readListFromDisk(d *disk.Disk, allIds *set.Set[model.BlockId], wg *sync.WaitGroup) {
	defer wg.Done()
	d.InListIds <- struct{}{}
	ids := <-d.OutListIds
	allIds.AddAll(&ids)
}
