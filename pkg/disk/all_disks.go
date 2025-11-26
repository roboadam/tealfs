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
)

type AllDisks struct {
	data   set.Set[model.AddDiskReq]
	loaded bool

	ConfigPath   string
	FileOps      FileOps
	OutDiskAdded chan<- model.AddDiskReq
}

func (a *AllDisks) Add(disk model.AddDiskReq) {
	a.load()
	a.data.Add(disk)
}

func (a *AllDisks) Get() set.Set[model.AddDiskReq] {
	a.load()
	return a.data.Clone()
}

func (a *AllDisks) load() {
	if a.loaded {
		return
	}
	a.loaded = true

	data, err := a.FileOps.ReadFile(filepath.Join(a.ConfigPath, "disks.json"))

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
