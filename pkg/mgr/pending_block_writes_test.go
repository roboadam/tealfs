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

package mgr

import (
	"tealfs/pkg/model"
	"testing"
)

func TestPendingBlockWrites(t *testing.T) {
	pbw := newPendingBlockWrites()
	blockId1 := model.PutBlockId("putBlockId1")
	blockId2 := model.PutBlockId("putBlockId2")
	nodeId := model.NewNodeId()
	ptr1 := model.NewDiskPointer(nodeId, "disk1", "someFile1")
	ptr2 := model.NewDiskPointer(nodeId, "disk1", "someFile2")
	ptr3 := model.NewDiskPointer(nodeId, "disk1", "someFile3")

	result := pbw.resolve(ptr1, blockId1)
	if result != notTracking {
		t.Errorf("should be not tracking")
		return
	}

	pbw.add(blockId1, ptr1)
	pbw.add(blockId1, ptr2)
	pbw.add(blockId2, ptr3)

	result = pbw.resolve(ptr1, blockId1)
	if result != notDone {
		t.Errorf("should not be done")
		return
	}

	result = pbw.resolve(ptr2, blockId1)
	if result != done {
		t.Errorf("should be done")
		return
	}

	result = pbw.resolve(ptr3, blockId2)
	if result != done {
		t.Errorf("should nbe done")
		return
	}
}
