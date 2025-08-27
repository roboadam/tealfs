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

package rebalancer

import (
	"context"
	"tealfs/pkg/model"
	"tealfs/pkg/set"
)

type DeleteUnused struct {
	InRunCleanup    <-chan AllBlockId
	OutRemoteDelete chan<- model.MgrConnsSend
	OutLocalDelete  chan<- DeleteBlockId

	OnDiskIds       *set.Map[AllBlockId, AllBlockIdResp]
	OnFilesystemIds *set.Map[AllBlockId, AllBlockIdResp]
	Mapper          *model.NodeConnectionMapper
	NodeId          model.NodeId
}

func (d *DeleteUnused) Start(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			return
		case req := <-d.InRunCleanup:
			onDisk := d.OnDiskSet(req)
			active := d.ActiveSet(req)
			toDelete := onDisk.Minus(active)
			for _, blockId := range toDelete.GetValues() {
				deleteMsg := DeleteBlockId{BlockId: blockId}
				d.OutLocalDelete <- deleteMsg

				connections := d.Mapper.Connections()
				for _, connId := range connections.GetValues() {
					d.OutRemoteDelete <- model.MgrConnsSend{
						Payload: &deleteMsg,
						ConnId:  connId,
					}
				}
			}
		}
	}
}

func (d *DeleteUnused) OnDiskSet(key AllBlockId) *set.Set[model.BlockId] {
	return setForResp(key, d.OnDiskIds)
}

func (d *DeleteUnused) ActiveSet(key AllBlockId) *set.Set[model.BlockId] {
	return setForResp(key, d.OnFilesystemIds)
}

func setForResp(key AllBlockId, blockMap *set.Map[AllBlockId, AllBlockIdResp]) *set.Set[model.BlockId] {
	resp, ok := blockMap.Get(key)
	if !ok {
		panic("don't know what to save")
	}
	return &resp.BlockIds
}
