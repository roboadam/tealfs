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
	"tealfs/pkg/disk"
	"tealfs/pkg/disk/dist"
	"tealfs/pkg/model"
	"tealfs/pkg/set"

	"github.com/google/uuid"
)

type SaveGetLocally struct {
	InGet <-chan GetFromDiskResp

	LocalDisks *set.Set[disk.Disk]
	Dist       *dist.MirrorDistributer
	NodeId     model.NodeId
}

func (s *SaveGetLocally) Start(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			return
		case resp := <-s.InGet:
			saveToIds := s.saveToDiskIds(&resp)
			for _, disk := range s.LocalDisks.GetValues() {
				if saveToIds.Contains(disk.Id()) {
					wr := s.writeRequest(disk, resp)
					disk.InWrites <- wr
				}
			}
		}
	}
}

func (s *SaveGetLocally) writeRequest(disk disk.Disk, resp GetFromDiskResp) model.WriteRequest {
	wr := model.WriteRequest{
		Caller: s.NodeId,
		Data: model.RawData{
			Ptr: model.DiskPointer{
				NodeId:   s.NodeId,
				Disk:     disk.Id(),
				FileName: string(resp.Resp.Block.Id),
			},
			Data: resp.Resp.Block.Data,
		},
		ReqId: model.PutBlockId(uuid.NewString()),
	}
	return wr
}

func (s *SaveGetLocally) saveToDiskIds(resp *GetFromDiskResp) set.Set[model.DiskId] {
	result := set.NewSet[model.DiskId]()
	blockId := resp.Resp.Block.Id
	pointers := s.Dist.WritePointersForId(blockId)
	for _, pointer := range pointers {
		result.Add(pointer.Disk)
	}
	return result
}
