// Copyright (C) 2026 Adam Hess
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

package datalayer

import (
	"context"
	"tealfs/pkg/disk"
	"tealfs/pkg/model"
	"tealfs/pkg/set"
)

type DeleteRequestHandler struct {
	InDeleteRequests <-chan DeleteRequest
	Disks            *set.Set[disk.Disk]
	NodeId           model.NodeId
	StateHandler     *StateHandler
}

func (d *DeleteRequestHandler) Start(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			return
		case req := <-d.InDeleteRequests:
			d.handleDeleteRequest(req)
		}
	}
}

func (d *DeleteRequestHandler) handleDeleteRequest(req DeleteRequest) {
	for _, disk := range d.Disks.GetValues() {
		if req.Dest.DiskId == disk.DiskId() {
			// disk.Delete
		}
	}
}
