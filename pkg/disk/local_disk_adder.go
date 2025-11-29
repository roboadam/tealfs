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
	"tealfs/pkg/disk/dist"
	"tealfs/pkg/model"
	"tealfs/pkg/set"
)

type LocalDiskAdder struct {
	InAddDiskReq     <-chan model.AddDiskReq
	OutAddLocalDisk  []chan<- *Disk
	OutIamDiskUpdate chan<- []model.AddDiskReq

	FileOps     FileOps
	Disks       *set.Set[Disk]
	Distributer *dist.MirrorDistributer
	AllDiskIds  *AllDisks
}

func (l *LocalDiskAdder) Start(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			return
		case add := <-l.InAddDiskReq:
			path := NewPath(add.Path, l.FileOps)
			disk := New(path, add.NodeId, add.DiskId, ctx)

			l.Disks.Add(disk)
			l.Distributer.SetWeight(add.NodeId, add.DiskId, 1)
			l.AllDiskIds.Add(add)

			for _, diskChan := range l.OutAddLocalDisk {
				diskChan <- &disk
			}

			all := l.AllDiskIds.Get()
			l.OutIamDiskUpdate <- all.GetValues()
		}
	}
}
