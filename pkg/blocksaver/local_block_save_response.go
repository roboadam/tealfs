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
	"tealfs/pkg/set"
)

type LocalBlockSaveResponses struct {
	Disks          *set.Set[disk.Disk]
	WriteResponses chan<- SaveToDiskResp
}

func (l *LocalBlockSaveResponses) Start(ctx context.Context) {
	for _, disk := range l.Disks.GetValues() {
		go l.readFromChan(ctx, disk.OutWrites)
	}
}

func (l *LocalBlockSaveResponses) readFromChan(ctx context.Context, c <-chan model.WriteResult) {
	for {
		select {
		case <-ctx.Done():
			return
		case wr := <-c:
			l.WriteResponses <- *convert(&wr)
		}
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
			Disk:   wr.Ptr.Disk,
		},
		Resp: model.PutBlockResp{
			Id:  wr.ReqId,
			Err: err,
		},
	}
}
