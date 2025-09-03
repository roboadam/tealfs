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

	"github.com/google/uuid"
)

type FillMissing struct {
	InStart  <-chan AllBlockId
	OutSends chan<- model.MgrConnsSend

	OnFilesystemIds *set.Map[AllBlockId, AllBlockIdResp]
	Mapper          *model.NodeConnectionMapper
	NodeId          model.NodeId

	requestMap map[AllBlockId]map[StoreItId]bool
}

func (f *FillMissing) Run(ctx context.Context) {
	f.requestMap = make(map[AllBlockId]map[StoreItId]bool)
	for {
		select {
		case <-ctx.Done():
			return
		case rebalanceId := <-f.InStart:
			blockIdsHolder, ok := f.OnFilesystemIds.Get(rebalanceId)
			if ok {
				blockIds := blockIdsHolder.BlockIds
				f.requestMap[rebalanceId] = make(map[StoreItId]bool)
				for _, blockId := range blockIds.GetValues() {
					storeIdId := StoreItId(uuid.New().String())
					f.requestMap[rebalanceId][storeIdId] = true
					f.OutSends <- model.MgrConnsSend{
						Payload: &StoreItCmd{
							StoreItId: storeIdId,
							BlockId:   blockId,
							Caller:    f.NodeId,
						},
						ConnId: 0,
					}
				}
			}
		}
	}
}
