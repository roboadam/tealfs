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

type Disks struct {
	Distributer dist.MirrorDistributer
	AllDiskIds  set.Set[model.DiskIdPath]
	Disks       set.Set[Disk]

	AddLocalDisk  <-chan AddDiskReq
	AddRemoteDisk <-chan AddRemoteDiskReq
	AddedDisk     []chan<- *Disk
}

func NewDisks(nodeId model.NodeId) *Disks {
	distributer := dist.NewMirrorDistributer(nodeId)
	allDiskIds := set.NewSet[model.DiskIdPath]()
	disks := set.NewSet[Disk]()
	return &Disks{
		Distributer: distributer,
		AllDiskIds:  allDiskIds,
		Disks:       disks,
	}
}

type AddRemoteDiskReq struct {
	NodeId model.NodeId
	DiskId model.DiskId
	Path   Path
}

func (d *Disks) Start(ctx context.Context) {
	select {
	case <-ctx.Done():
		return
	case add := <-d.AddLocalDisk:
		disk := New(add.Path, add.Id, add.DiskId, ctx)
		d.Disks.Add(disk)
		d.Distributer.SetWeight(add.Id, add.DiskId, 1)
		d.AllDiskIds.Add(model.DiskIdPath{
			Id:   add.DiskId,
			Path: add.Path.String(),
			Node: add.Id,
		})
		for _, diskChan := range d.AddedDisk {
			diskChan <- &disk
		}

	case add := <-d.AddRemoteDisk:
		d.Distributer.SetWeight(add.NodeId, add.DiskId, 1)
		d.AllDiskIds.Add(model.DiskIdPath{
			Id:   add.DiskId,
			Path: add.Path.String(),
			Node: add.NodeId,
		})
	}
}
