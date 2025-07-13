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

package blocksaver

import (
	"context"
	"tealfs/pkg/disk/dist"
	"tealfs/pkg/model"
)

type BlockSaver struct {
	Req         <-chan model.PutBlockReq
	RemoteDest  chan<- SaveToDiskReq
	LocalDest   chan<- SaveToDiskReq
	Distributer *dist.MirrorDistributer
	NodeId      model.NodeId
}

type Dest struct {
	NodeId model.NodeId
	Disk   model.DiskId
}

type SaveToDiskReq struct {
	Caller model.NodeId
	Dest   Dest
	Req    model.PutBlockReq
}

type SaveToDiskResp struct {
	Caller model.NodeId
	Dest   Dest
	Resp   model.PutBlockResp
}

func (bs *BlockSaver) Start(ctx context.Context) {
	for {
		select {
		case req := <-bs.Req:
			dests := bs.destsFor(req)
			for _, dest := range dests {
				saveToDisk := SaveToDiskReq{Dest: dest, Req: req}
				if dest.NodeId == bs.NodeId {
					bs.LocalDest <- saveToDisk
				} else {
					bs.RemoteDest <- saveToDisk
				}
			}
		case <-ctx.Done():
			return
		}
	}
}
