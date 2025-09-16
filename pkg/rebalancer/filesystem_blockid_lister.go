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
	"tealfs/pkg/webdav"
)

type ActiveBlockIdLister struct {
	InFetchIds      <-chan ListOnDiskBlockIdsCmd
	OutLocalResults chan<- FilesystemBlockIdList

	FileSystem *webdav.FileSystem
}

func (l *ActiveBlockIdLister) Start(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			return
		case req := <-l.InFetchIds:
			l.collectResults(req)
		}
	}
}

func (l *ActiveBlockIdLister) collectResults(req ListOnDiskBlockIdsCmd) {
	allIds := l.FileSystem.ListBlockIds()
	l.OutLocalResults <- FilesystemBlockIdList{
		Caller:       req.Caller,
		BlockIds:     *allIds,
		BalanceReqId: req.BalanceReqId,
	}
}
