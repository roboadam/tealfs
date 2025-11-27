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
	"encoding/json"
	"errors"
	"io/fs"
	"path/filepath"
	"tealfs/pkg/model"
	"tealfs/pkg/set"

	log "github.com/sirupsen/logrus"
)

type AllDisks struct {
	data       set.Set[model.AddDiskReq]
	configPath string
	fileOps    FileOps

	OutDiskAdded chan<- model.AddDiskReq
}

func (a *AllDisks) Add(disk model.AddDiskReq) {
	added := a.data.Add(disk)
	if added {
		data, err := json.Marshal(a.data.GetValues())
		if err != nil {
			log.Panicf("Error saving disk ids %v", err)
		}

		err = a.fileOps.WriteFile(filepath.Join(a.configPath, "disks.json"), data)
		if err != nil {
			log.Panicf("Error saving disk ids %v", err)
		}

		a.OutDiskAdded <- disk
	}
}

func (a *AllDisks) Get() set.Set[model.AddDiskReq] {
	return a.data.Clone()
}

func (a *AllDisks) Init(configPath string, fileOps FileOps) {
	a.configPath = configPath
	a.fileOps = fileOps

	data, err := fileOps.ReadFile(filepath.Join(configPath, "disks.json"))

	if errors.Is(err, fs.ErrNotExist) {
		a.data = set.NewSet[model.AddDiskReq]()
		return
	}

	diskInfo := []model.AddDiskReq{}
	if err == nil {
		err = json.Unmarshal(data, &diskInfo)
		if err == nil {
			a.data = set.NewSetFromSlice(diskInfo)
			a.sendAllDiskAdded()
		}
	}
}

func (a *AllDisks) sendAllDiskAdded() {
	for _, dip := range a.data.GetValues() {
		a.OutDiskAdded <- dip
	}
}
