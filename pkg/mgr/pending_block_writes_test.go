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
	blockId1 := model.NewBlockId()
	blockId2 := model.NewBlockId()
	nodeId := model.NewNodeId()
	ptr1 := model.DiskPointer{
		NodeId:   nodeId,
		FileName: "someFile1",
	}
	ptr2 := model.DiskPointer{
		NodeId:   nodeId,
		FileName: "someFile2",
	}
	ptr3 := model.DiskPointer{
		NodeId:   nodeId,
		FileName: "someFile3",
	}

	result, _ := pbw.resolve(ptr1)
	if result != notTracking {
		t.Errorf("should be not tracking")
		return
	}

	pbw.add(blockId1, ptr1)
	pbw.add(blockId1, ptr2)
	pbw.add(blockId2, ptr3)

	result, blockResult := pbw.resolve(ptr1)
	if result != notDone || blockResult != blockId1 {
		t.Errorf("should not be done")
		return
	}

	result, blockResult = pbw.resolve(ptr2)
	if result != done || blockResult != blockId1 {
		t.Errorf("should be done")
		return
	}

	result, blockResult = pbw.resolve(ptr3)
	if result != done || blockResult != blockId2 {
		t.Errorf("should nbe done")
		return
	}
}
