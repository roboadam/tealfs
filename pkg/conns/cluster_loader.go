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
	"errors"
	"io/fs"
	"path/filepath"
	"tealfs/pkg/disk"
	"tealfs/pkg/model"
	"time"

	"github.com/sirupsen/logrus"
)

type ClusterLoader struct {
	NodeConnMapper *model.NodeConnectionMapper
	FileOps        disk.FileOps
	SavePath       string
}

func (c *ClusterLoader) Load(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			return
		default:
			err := c.loadSettings()
			if err == nil {
				return
			} else {
				logrus.Warnf("unable to load cluster.json: %v", err)
				time.Sleep(time.Millisecond * 500)
			}
		}
	}
}

func (c *ClusterLoader) loadSettings() error {
	data, err := c.FileOps.ReadFile(filepath.Join(c.SavePath, "cluster.json"))
	if err != nil && !errors.Is(err, fs.ErrNotExist) {
		return err
	}
	if len(data) > 0 {
		mapper, err := model.NodeConnectionMapperUnmarshal(data)
		if err != nil {
			return err
		}
		c.NodeConnMapper.Clear()
		for _, na := range mapper.NodesWithAddress() {
			c.NodeConnMapper.SetNodeAddress(na.J, na.K)
		}
	}

	return nil
}
