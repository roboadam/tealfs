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
	"log"
	"path/filepath"
	"sort"
	"tealfs/pkg/chanutil"
	"tealfs/pkg/disk"
	"tealfs/pkg/disk/dist"
	"tealfs/pkg/model"
	"tealfs/pkg/set"
	"time"
)

type Custodian struct {
	ctx               context.Context
	nodeId            model.NodeId
	nodes             set.Set[model.NodeId]
	globalBlockIds    set.Set[model.BlockId]
	verifyBlockId     chan model.BlockId
	fileOps           disk.FileOps
	savePath          string
	inBroadcast       chan model.Broadcast
	mirrorDistributer *dist.MirrorDistributer
	ConnsSends        chan model.MgrConnsSend
}

func New(ctx context.Context, nodeId model.NodeId, nodes set.Set[model.NodeId], fileOps disk.FileOps, savePath string) (*Custodian, error) {
	c := Custodian{
		ctx:           ctx,
		nodeId:        nodeId,
		nodes:         nodes,
		verifyBlockId: make(chan model.BlockId),
		fileOps:       fileOps,
		savePath:      savePath,
	}

	if err := c.loadGbl(); err != nil {
		return nil, err
	}

	go c.verifyGlobalBlockListIfMain()
	go c.mainLoop()

	return &c, nil
}

func (c *Custodian) mainNodeId() model.NodeId {
	values := c.nodes.GetValues()
	sort.Slice(values, func(i, j int) bool {
		return values[i] < values[j]
	})
	return values[0]
}

func (c *Custodian) mainLoop() {
	for {
		select {
		case <-c.ctx.Done():
			return
		case b := <-c.inBroadcast:
			c.handleBroadcast(b)
		}
	}
}

func (c *Custodian) handleBroadcast(b model.Broadcast) {
	bCast := MgrBroadcastMsgFromBytes(b.Msg())
	if bCast.GBList == nil {
		cmd := bCast.GBLCmd
		switch cmd.Type {
		case Add:
			c.globalBlockIds.Add(cmd.BlockId)
			err := c.saveGbl()
			if err != nil {
				log.Panicf("%v", err)
			}
		case Delete:
			c.globalBlockIds.Remove(cmd.BlockId)
			err := c.saveGbl()
			if err != nil {
				log.Panicf("%v", err)
			}
		}
	} else {
		c.globalBlockIds = *bCast.GBList
	}
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
	path := filepath.Join(c.savePath, "gbl.bin")
	return SaveGBL(c.fileOps, path, &c.globalBlockIds)
}

func (c *Custodian) loadGbl() error {
	path := filepath.Join(c.savePath, "gbl.bin")
	gbl, err := LoadGBL(c.fileOps, path)
	if err != nil {
		return err
	}
	c.globalBlockIds = *gbl
	return nil
}

func (c *Custodian) handleVerifyBlockId(id model.BlockId) {
	disks := c.mirrorDistributer.ReadPointersForId(id)
	if len(disks) == 0 {
		return
	}
	for _, disk := range disks {
		if disk.NodeId() != c.nodeId {
			// send new has block payload to dest node
		}
	}
}

func (c *Custodian) reconcileBlocks() {
	if c.mainNodeId() == c.nodeId {
		mgrBroadcastMsg := MgrBroadcastMsg{GBList: &c.globalBlockIds}
		broadcast := model.NewBroadcast(mgrBroadcastMsg.ToBytes(), model.CustodianDest)
		// Todo: make sure blocks aren't added before this message goes out
		c.ConnsSends <- model.MgrConnsSend{Payload: &broadcast}
	}
}
