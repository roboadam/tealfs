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
)

type LocalBlockSaver struct {
	Req   <-chan SaveToDiskReq
	Disks []disk.Disk
}

func (l *LocalBlockSaver) Start(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			return
		case req := <-l.Req:
			if disk, err := l.diskForId(req.Dest.Disk); err == nil {
				disk.InWrites <- model.WriteRequest{
					Caller: req.Caller,
					Data: model.RawData{
						Ptr:  model.DiskPointer{},
						Data: []byte{},
					},
					ReqId: req.Req.Id(),
				}
			}
		}
	}
}

func ConvertSaveReq(req *SaveToDiskReq) *model.WriteRequest {
	return &model.WriteRequest{
		Caller: req.Caller,
		Data: model.RawData{
			Ptr: model.DiskPointer{
				NodeId:   req.Dest.NodeId,
				Disk:     req.Dest.Disk,
				FileName: string(req.Req.Block.Id),
			},
			Data: req.Req.Block.Data,
		},
		ReqId: req.Req.Id(),
	}
}

func (l *LocalBlockSaver) diskForId(diskId model.DiskId) (*disk.Disk, error) {
	for _, disk := range l.Disks {
		if disk.Id() == diskId {
			return &disk, nil
		}
	}
	return nil, errors.New("disk not found")
}
