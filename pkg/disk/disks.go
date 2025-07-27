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
	NodeId      model.NodeId
	FileOps     FileOps

	InAddDisk       <-chan model.DiskIdPath
	OutAddLocalDisk []chan<- *Disk
	OutStatus       chan<- model.UiDiskStatus
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

func (d *Disks) Start(ctx context.Context) {
	select {
	case <-ctx.Done():
		return
	case add := <-d.InAddDisk:
		if !d.AllDiskIds.Contains(add) {
			d.addDisk(ctx, add)
		}
	}
}

func (d *Disks) addDisk(ctx context.Context, add model.DiskIdPath) {
	if add.Node == d.NodeId {
		d.addLocalDisk(ctx, add)
	} else {
		d.addRemoteDisk(add)
	}
}

func (d *Disks) addLocalDisk(ctx context.Context, add model.DiskIdPath) {
	path := NewPath(add.Path, d.FileOps)
	disk := New(path, add.Node, add.Id, ctx)

	d.Disks.Add(disk)
	d.Distributer.SetWeight(add.Node, add.Id, 1)
	d.AllDiskIds.Add(add)

	for _, diskChan := range d.OutAddLocalDisk {
		diskChan <- &disk
	}
}

func (d *Disks) addRemoteDisk(add model.DiskIdPath) {
	d.Distributer.SetWeight(add.Node, add.Id, 1)
	d.AllDiskIds.Add(add)
}
