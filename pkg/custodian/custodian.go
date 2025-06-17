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

package custodian

import (
	"context"
	"path/filepath"
	"sort"
	"tealfs/pkg/chanutil"
	"tealfs/pkg/model"
	"tealfs/pkg/set"
	"time"
)

type Custodian struct {
	ctx            context.Context
	nodeId         model.NodeId
	nodes          set.Set[model.NodeId]
	globalBlockIds set.Set[model.BlockId]
	verifyBlockId  chan model.BlockId
}

func New(ctx context.Context, nodeId model.NodeId, nodes set.Set[model.NodeId], globalBlockIds set.Set[model.BlockId]) *Custodian {
	c := Custodian{
		ctx:            ctx,
		nodeId:         nodeId,
		nodes:          nodes,
		globalBlockIds: globalBlockIds,
		verifyBlockId:  make(chan model.BlockId),
	}
	go c.verifyGlobalBlockListIfMain()

	return &c
}

func (c *Custodian) mainNodeId() model.NodeId {
	values := c.nodes.GetValues()
	sort.Slice(values, func(i, j int) bool {
		return values[i] < values[j]
	})
	return values[0]
}

func (c *Custodian) verifyGlobalBlockListIfMain() {
	for {
		select {
		case <-c.ctx.Done():
			return
		default:
			if c.mainNodeId() != c.nodeId {
				time.Sleep(time.Hour)
			}
			for _, blockId := range c.globalBlockIds.GetValues() {
				chanutil.Send(c.ctx, c.verifyBlockId, blockId, "send verify")
				time.Sleep(time.Minute)
			}
		}
	}
}

func (c *Custodian) saveGbl() error {
	path := filepath.Join(m.savePath, "gbl.bin")
	return SaveGBL(c.fileOps, path, &c.globalBlockIds)
}

func (c *Custodian) loadGbl() error {
	path := filepath.Join(m.savePath, "gbl.bin")
	gbl, err := LoadGBL(m.fileOps, path)
	if err != nil {
		return err
	}
	m.GlobalBlockIds = *gbl
	return nil
}
