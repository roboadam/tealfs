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

package disk_test

import (
	"context"
	"tealfs/pkg/disk"
	"tealfs/pkg/set"
	"testing"
)

func TestDeleteBlocks(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	inDelete := make(chan disk.DeleteBlockId)
	disks := set.NewSet[disk.Disk]()

	deleteBlocks := disk.DeleteBlocks{
		InDelete: inDelete,
		Disks:    &disks,
	}
	go deleteBlocks.Start(ctx)

	ops := disk.MockFileOps{DataRemoved: make(chan struct{})}
	p := disk.NewPath("", &ops)
	d := disk.New(p, "nodeId", "diskId", ctx)
	ops.WriteFile("blockId1", []byte{1, 2, 3, 4, 5})
	disks.Add(d)

	inDelete <- disk.DeleteBlockId{
		BlockId: "blockId1",
	}

	<-ops.DataRemoved

	if ops.Exists("blockId1") {
		t.Error("That block should have been deleted")
	}
}
