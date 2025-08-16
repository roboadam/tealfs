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

package blockreader

import (
	"context"
	"errors"
	"tealfs/pkg/disk"
	"tealfs/pkg/model"

	log "github.com/sirupsen/logrus"
)

type LocalBlockReadResponses struct {
	InDisks            <-chan *disk.Disk
	LocalReadResponses chan<- GetFromDiskResp
	Sends              chan<- model.MgrConnsSend
	NodeConnMap        *model.NodeConnectionMapper
	NodeId             model.NodeId
}

func (l *LocalBlockReadResponses) Start(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			return
		case d := <-l.InDisks:
			go l.readFromChan(ctx, d.OutReads)
		}
	}
}

func (l *LocalBlockReadResponses) readFromChan(ctx context.Context, c <-chan model.ReadResult) {
	for {
		select {
		case <-ctx.Done():
			return
		case wr := <-c:
			resp := convert(&wr)
			if resp.Caller == l.NodeId {
				l.LocalReadResponses <- *convert(&wr)
			} else {
				l.sendToRemote(resp)
			}
		}
	}
}

func (l *LocalBlockReadResponses) sendToRemote(resp *GetFromDiskResp) {
	conn, ok := l.NodeConnMap.ConnForNode(resp.Caller)
	if ok {
		l.Sends <- model.MgrConnsSend{
			ConnId:  conn,
			Payload: resp,
		}
	} else {
		log.Warn("lbsr no connection")
	}
}

func convert(rr *model.ReadResult) *GetFromDiskResp {
	var err error
	var block model.Block
	if rr.Ok {
		block = model.Block{
			Id:   rr.BlockId,
			Data: rr.Data.Data,
		}
	} else {
		err = errors.New(rr.Message)
	}
	return &GetFromDiskResp{
		Caller: rr.Caller,
		Dest: Dest{
			NodeId: rr.Data.Ptr.NodeId,
			DiskId: rr.Data.Ptr.Disk,
		},
		Resp: model.GetBlockResp{
			Id:    rr.ReqId,
			Block: block,
			Err:   err,
		},
	}
}
