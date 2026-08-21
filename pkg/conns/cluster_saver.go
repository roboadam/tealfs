// Copyright (C) 2026 Adam Hess
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

package conns

import (
	"path/filepath"
	"sync"
	"tealfs/pkg/disk"
	"tealfs/pkg/model"
)

type ClusterSaver struct {
	NodeConnMapper *model.NodeConnectionMapper
	SavePath       string
	FileOps        disk.FileOps
	mux            sync.Mutex
}

func (c *ClusterSaver) Save() error {
	data, err := c.NodeConnMapper.Marshal()
	if err != nil {
		return err
	}

	c.mux.Lock()
	defer c.mux.Unlock()
	err = c.FileOps.WriteFile(filepath.Join(c.SavePath, "cluster.json"), data)
	if err != nil {
		return err
	}

	return nil
}
