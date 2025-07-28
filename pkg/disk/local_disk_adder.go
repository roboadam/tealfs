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
	AllDiskIds  *set.Set[model.AddDiskReq]
}

func (l *LocalDiskAdder) Start(ctx context.Context) {
	select {
	case <-ctx.Done():
		return
	case add := <-l.InAddDiskReq:
		path := NewPath(add.Path, l.FileOps)
		disk := New(path, add.Node, add.Id, ctx)

		l.Disks.Add(disk)
		l.Distributer.SetWeight(add.Node, add.Id, 1)
		l.AllDiskIds.Add(add)

		for _, diskChan := range l.OutAddLocalDisk {
			diskChan <- &disk
		}

		l.OutIamDiskUpdate <- l.AllDiskIds.GetValues()
	}
}
