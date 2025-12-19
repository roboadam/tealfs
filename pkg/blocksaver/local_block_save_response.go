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

package blocksaver

import (
	"context"
	"errors"
	"tealfs/pkg/disk"
	"tealfs/pkg/model"

	log "github.com/sirupsen/logrus"
)

type LocalBlockSaveResponses struct {
	InDisks             <-chan *disk.Disk
	LocalWriteResponses chan<- SaveToDiskResp
	Sends               chan<- model.SendPayloadMsg
	NodeConnMap         *model.NodeConnectionMapper
	NodeId              model.NodeId
}

func (l *LocalBlockSaveResponses) Start(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			return
		case d := <-l.InDisks:
			go l.readFromChan(ctx, d.OutWrites)
		}
	}
}

func (l *LocalBlockSaveResponses) readFromChan(ctx context.Context, c <-chan model.WriteResult) {
	for {
		select {
		case <-ctx.Done():
			return
		case wr := <-c:
			resp := convert(&wr)
			if resp.Caller == l.NodeId {
				l.LocalWriteResponses <- *convert(&wr)
			} else {
				l.sendToRemote(resp)
			}
		}
	}
}

func (l *LocalBlockSaveResponses) sendToRemote(resp *SaveToDiskResp) {
	conn, ok := l.NodeConnMap.ConnForNode(resp.Caller)
	if ok {
		l.Sends <- model.SendPayloadMsg{
			ConnId:  conn,
			Payload: resp,
		}
	} else {
		log.Warn("lbsr no connection")
	}
}

func convert(wr *model.WriteResult) *SaveToDiskResp {
	var err error
	if !wr.Ok {
		err = errors.New(wr.Message)
	}
	return &SaveToDiskResp{
		Caller: wr.Caller,
		Dest: Dest{
			NodeId: wr.Ptr.NodeId,
			DiskId: wr.Ptr.Disk,
		},
		Resp: model.PutBlockResp{
			Id:  wr.ReqId,
			Err: err,
		},
	}
}
