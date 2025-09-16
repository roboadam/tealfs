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

	"github.com/google/uuid"
)

type LocalStoreItHandler struct {
	InStoreItCmd   <-chan StoreItCmd
	OutGetBlockReq chan<- model.GetBlockReq
}

func (l *LocalStoreItHandler) Start(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			return
		case cmd := <-l.InStoreItCmd:
			// 1. Check if we already have the disk
			// 2. Get a list of Node+Disks that could have the data
			// 3. Save that list
			// 2. If not send it a GetBlockForRebalanceReq
			l.OutGetBlockReq <- model.GetBlockReq{
				Id:      model.GetBlockId(uuid.NewString()),
				BlockId: cmd.BlockId,
			}
		}
	}
}
