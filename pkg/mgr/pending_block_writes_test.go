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

	if pbw.resolve(ptr1) == false {
		t.Errorf("should be resolved")
		return
	}

	pbw.add(blockId1, ptr1)
	pbw.add(blockId1, ptr2)
	pbw.add(blockId2, ptr3)

	if pbw.resolve(ptr1) == true {
		t.Errorf("should not be resolved")
		return
	}

	if pbw.resolve(ptr1) == true {
		t.Errorf("should not be resolved")
		return
	}

	if pbw.resolve(ptr2) == false {
		t.Errorf("should be resolved")
		return
	}

	if pbw.resolve(ptr3) == false {
		t.Errorf("should be resolved")
		return
	}
}
