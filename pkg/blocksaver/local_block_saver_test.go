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
	"tealfs/pkg/disk"
	"tealfs/pkg/model"
	"tealfs/pkg/set"
	"testing"

	"github.com/google/uuid"
)

func TestLocalBlockSaver(t *testing.T) {
	nodeId := model.NewNodeId()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	req := make(chan SaveToDiskReq)

	disks := set.NewSet[disk.Disk]()
	disks.Add(*mockDisk(nodeId, ctx))
	disks.Add(*mockDisk(nodeId, ctx))

	lbs := LocalBlockSaver{
		Req:   req,
		Disks: &disks,
	}

	lbs.Start(ctx)
}

func mockDisk(nodeId model.NodeId, ctx context.Context) *disk.Disk {
	p := disk.NewPath("/test", &disk.MockFileOps{})
	d := disk.New(
		p,
		nodeId,
		model.DiskId(uuid.NewString()),
		make(chan model.WriteRequest),
		make(chan model.ReadRequest),
		make(chan model.WriteResult),
		make(chan model.ReadResult),
		ctx,
	)
	return &d
}
