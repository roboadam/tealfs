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

type DiskLoader struct {
	FileOps  FileOps
	SavePath string
}

func (d *DiskLoader) LoadDiskIds() (set.Set[model.DiskIdPath], error) {
	diskInfo := []model.DiskIdPath{}
	data, err := d.FileOps.ReadFile(filepath.Join(d.SavePath, "disks.json"))
	if err == nil {
		err = json.Unmarshal(data, &diskInfo)
		if err != nil {
			return set.NewSet[model.DiskIdPath](), err
		}
	} else if !errors.Is(err, fs.ErrNotExist) {
		return set.NewSet[model.DiskIdPath](), err
	}
	return set.NewSetFromSlice(diskInfo), nil
}
