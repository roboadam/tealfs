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

type DiskManagerSvc struct {
	Distributer dist.MirrorDistributer
	// AllDiskIds       *AllDisks
	DiskInfoList     set.Set[DiskInfo]
	LocalDiskSvcList set.Set[Disk]
	NodeId           model.NodeId

	InAddDiskMsg    <-chan model.AddDiskMsg
	OutDiskAddedMsg chan<- model.DiskAddedMsg

	configPath string
	fileOps    FileOps
}

type DiskInfo struct {
	NodeId model.NodeId
	DiskId model.DiskId
	Path   string
}

func NewDisks(nodeId model.NodeId, configPath string, fileOps FileOps) *DiskManagerSvc {
	distributer := dist.NewMirrorDistributer(nodeId)
	localDisks := set.NewSet[Disk]()
	diskInfoList := set.NewSet[DiskInfo]()
	return &DiskManagerSvc{
		Distributer:      distributer,
		DiskInfoList:     diskInfoList,
		LocalDiskSvcList: localDisks,
		NodeId:           nodeId,

		configPath: configPath,
		fileOps:    fileOps,
	}
}

func (d *DiskManagerSvc) Start(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			return
		case add := <-d.InAddDiskMsg:
			if !d.localDiskExists(add) {
				path := NewPath()
				disk := New(pa)
				// d.LocalDiskSvcList.Add()
			}
		}
	}
}

func (d *DiskManagerSvc) localDiskExists(add model.AddDiskMsg) bool {
	for _, disk := range d.LocalDiskSvcList.GetValues() {
		if disk.diskId == add.DiskId {
			return true
		}
	}
	return false
}
