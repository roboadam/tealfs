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
	AllDiskIds  set.Set[model.AddDiskReq]
	Disks       set.Set[Disk]
	NodeId      model.NodeId

	InAddDiskReq        <-chan model.AddDiskReq
	OutRemoteAddDiskReq chan<- model.AddDiskReq
	OutLocalAddDiskReq  chan<- model.AddDiskReq
}

func NewDisks(nodeId model.NodeId) *Disks {
	distributer := dist.NewMirrorDistributer(nodeId)
	allDiskIds := set.NewSet[model.AddDiskReq]()
	disks := set.NewSet[Disk]()
	return &Disks{
		Distributer: distributer,
		AllDiskIds:  allDiskIds,
		Disks:       disks,
		NodeId:      nodeId,
	}
}

func (d *Disks) Start(ctx context.Context) {
	select {
	case <-ctx.Done():
		return
	case add := <-d.InAddDiskReq:
		if !d.AllDiskIds.Contains(add) {
			d.addDisk(add)
		}
	}
}

func (d *Disks) addDisk(add model.AddDiskReq) {
	if add.NodeId == d.NodeId {
		d.OutLocalAddDiskReq <- add
	} else {
		d.OutRemoteAddDiskReq <- add
	}
}
