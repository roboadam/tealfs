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

	log "github.com/sirupsen/logrus"
)

type ActiveBlockIdLister struct {
	InFetchIds       <-chan BalanceReq
	OutLocalResults  chan<- FilesystemBlockIdList
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
	l.sendResults(&FilesystemBlockIdList{
		Caller:       req.Caller,
		BlockIds:     allIds,
		BalanceReqId: req.BalanceReqId,
	})
}

func (l *ActiveBlockIdLister) sendResults(resp *FilesystemBlockIdList) {
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
