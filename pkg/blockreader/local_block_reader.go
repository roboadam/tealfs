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
	"tealfs/pkg/set"

	log "github.com/sirupsen/logrus"
)

type LocalBlockReader struct {
	Req   <-chan GetFromDiskReq
	Disks *set.Set[disk.Disk]
}

func (l *LocalBlockReader) Start(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			return
		case req := <-l.Req:
			if disk, err := l.diskForId(req.Dest.DiskId); err == nil {
				disk.InReads <- *convertReadReq(&req)
			} else {
				log.Panicf("no disk for id %s, %v", req.Dest.DiskId, err)
			}
		}
	}
}

func convertReadReq(req *GetFromDiskReq) *model.ReadRequest {
	return &model.ReadRequest{
		Caller: req.Caller,
		ReqId:  req.Req.Id,
		Ptrs: []model.DiskPointer{{
			NodeId:   req.Caller,
			Disk:     req.Dest.DiskId,
			FileName: string(req.Req.BlockId),
		}},
		BlockId: req.Req.BlockId,
	}
}

func (l *LocalBlockReader) diskForId(diskId model.DiskId) (*disk.Disk, error) {
	for _, disk := range l.Disks.GetValues() {
		if disk.Id() == diskId {
			return &disk, nil
		}
	}
	return nil, errors.New("disk not found")
}
