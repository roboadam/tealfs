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
	"encoding/json"
	"errors"
	"io/fs"
	"path/filepath"
	"tealfs/pkg/disk/dist"
	"tealfs/pkg/model"
	"tealfs/pkg/set"

	log "github.com/sirupsen/logrus"
)

type DiskManagerSvc struct {
	Distributer      dist.MirrorDistributer
	DiskInfoList     set.Set[DiskInfo]
	LocalDiskSvcList set.Set[Disk]
	NodeId           model.NodeId

	InAddDiskMsg    <-chan model.AddDiskMsg
	InDiskAddedMsg  <-chan model.DiskAddedMsg
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
	d.loadDiskInfoList(ctx)
	for {
		select {
		case <-ctx.Done():
			return
		case added:= <- d.InDiskAddedMsg:
			d.addToDiskInfoList(model.AddDiskMsg(added))
		case add := <-d.InAddDiskMsg:
			if !d.localDiskExists(add) {
				path := NewPath(d.configPath, d.fileOps)
				disk := New(path, d.NodeId, add.DiskId, ctx)
				added := d.LocalDiskSvcList.Add(disk)
				if added {
					d.OutDiskAddedMsg <- model.DiskAddedMsg(add)
				}
			}
			d.addToDiskInfoList(add)
		}
	}
}

func (d *DiskManagerSvc) addToDiskInfoList(add model.AddDiskMsg) {
	added := d.DiskInfoList.Add(DiskInfo{
		NodeId: add.NodeId,
		DiskId: add.DiskId,
		Path:   add.Path,
	})
	if added {
		d.Distributer.SetWeight(add.NodeId, add.DiskId, 1)
		d.saveDiskInfoList()
	}
}

func (d *DiskManagerSvc) loadDiskInfoList(ctx context.Context) {
	data, err := d.fileOps.ReadFile(filepath.Join(d.configPath, "disks.json"))

	if errors.Is(err, fs.ErrNotExist) {
		d.DiskInfoList = set.NewSet[DiskInfo]()
		return
	}

	diskInfo := []DiskInfo{}
	if err == nil {
		err = json.Unmarshal(data, &diskInfo)
		if err == nil {
			d.DiskInfoList = set.NewSetFromSlice(diskInfo)
			for _, dInfo := range diskInfo {
				if dInfo.NodeId == d.NodeId {
					path := NewPath(d.configPath, d.fileOps)
					disk := New(path, d.NodeId, dInfo.DiskId, ctx)
					d.LocalDiskSvcList.Add(disk)
				}
			}
		}
	}
}

func (d *DiskManagerSvc) saveDiskInfoList() {
	data, err := json.Marshal(d.DiskInfoList.GetValues())
	if err != nil {
		log.Panicf("Error saving disk ids %v", err)
	}

	err = d.fileOps.WriteFile(filepath.Join(d.configPath, "disks.json"), data)
	if err != nil {
		log.Panicf("Error saving disk ids %v", err)
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
