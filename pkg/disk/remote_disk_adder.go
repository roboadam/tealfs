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

	log "github.com/sirupsen/logrus"
)

type RemoteDiskAdder struct {
	InAddDiskReq <-chan model.AddDiskReq
	OutSends     chan<- model.MgrConnsSend
	NodeConnMap  *model.NodeConnectionMapper
}

func (r *RemoteDiskAdder) Start(ctx context.Context) {
	select {
	case <-ctx.Done():
		return
	case add := <-r.InAddDiskReq:
		connId, ok := r.NodeConnMap.ConnForNode(add.Node)
		if ok {
			r.OutSends <- model.MgrConnsSend{
				ConnId:  connId,
				Payload: &add,
			}
		} else {
			log.Warn("No connection")
		}
	}
}
