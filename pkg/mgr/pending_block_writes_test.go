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
