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
	InStart      <-chan BalanceReqId
	OutExistsReq chan<- ExistsReq

	OnFilesystemIds *set.Map[BalanceReqId, OnDiskBlockIdList]
	NodeId          model.NodeId
}

func (f *FillMissing) Run(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			return
		case rebalanceId := <-f.InStart:
			f.StartExistsReqs(rebalanceId)
		}
	}
}

func (f *FillMissing) StartExistsReqs(rebalanceId BalanceReqId) {
	list, ok := f.OnFilesystemIds.Get(rebalanceId)
	if ok {
		blockIds := list.BlockIds.GetValues()
		for _, blockId := range blockIds {
			f.OutExistsReq <- ExistsReq{
				BalanceReqId: rebalanceId,
				ExistsId:     ExistsId(uuid.NewString()),
				BlockId:      blockId,
				Caller:       f.NodeId,
			}
		}
	}
}
