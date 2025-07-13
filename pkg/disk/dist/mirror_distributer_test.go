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

package dist_test

import (
	"tealfs/pkg/disk/dist"
	"tealfs/pkg/model"
	"tealfs/pkg/set"
	"testing"
)

func TestMirror(t *testing.T) {
	d := dist.NewMirrorDistributer()
	node1 := model.NewNodeId()
	disk1 := model.DiskId("disk1")
	node2 := model.NewNodeId()
	disk2 := model.DiskId("disk2")
	node3 := model.NewNodeId()
	disk3 := model.DiskId("disk3")
	allNodes := set.NewSet[model.NodeId]()
	allNodes.Add(node1)
	allNodes.Add(node2)
	allNodes.Add(node3)

	d.SetWeight(node1, disk1, 1)
	d.SetWeight(node2, disk2, 2)
	d.SetWeight(node3, disk3, 4)

	bucket1 := 0
	bucket2 := 0
	bucket3 := 0

	for range 100 {
		nodes := allNodes.Clone()
		blockId := model.NewBlockId()
		ptrs := d.ReadPointersForId(blockId)

		if len(ptrs) != 3 {
			t.Error("should have 3 main data nodes had", len(ptrs))
			return
		}

		if !nodes.Contains(ptrs[0].NodeId) {
			t.Error("missing one of the nodes")
			return
		}
		nodes.Remove(ptrs[0].NodeId)

		if !nodes.Contains(ptrs[1].NodeId) {
			t.Error("missing one of the nodes")
			return
		}

		switch ptrs[0].NodeId {
		case node1:
			bucket1++
		case node2:
			bucket2++
		case node3:
			bucket3++
		}

		switch ptrs[1].NodeId {
		case node1:
			bucket1++
		case node2:
			bucket2++
		case node3:
			bucket3++
		}
	}

	if bucket1+bucket2+bucket3 != 200 {
		t.Error("should have 200 blocks")
		return
	}

	if bucket1 > bucket2 || bucket2 > bucket3 || bucket1 > bucket3 {
		t.Error("should be distributed")
		return
	}
}
