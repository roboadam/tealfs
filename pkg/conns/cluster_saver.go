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

package conns

import (
	"context"
	"path/filepath"
	"tealfs/pkg/disk"
	"tealfs/pkg/model"
)

type ClusterSaver struct {
	Save           <-chan struct{}
	NodeConnMapper *model.NodeConnectionMapper
	SavePath       string
	FileOps        disk.FileOps
}

func (c *ClusterSaver) Start(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			return
		case <-c.Save:
			c.save()
		}
	}
}

func (c *ClusterSaver) save() error {
	data, err := c.NodeConnMapper.Marshal()
	if err != nil {
		return err
	}

	err = c.FileOps.WriteFile(filepath.Join(c.SavePath, "cluster.json"), data)
	if err != nil {
		return err
	}

	return nil
}
