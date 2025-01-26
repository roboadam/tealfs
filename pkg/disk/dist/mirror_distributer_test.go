// Copyright (C) 2024 Adam Hess
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
	node2 := model.NewNodeId()
	node3 := model.NewNodeId()
	allNodes := set.NewSet[model.NodeId]()
	allNodes.Add(node1)
	allNodes.Add(node2)
	allNodes.Add(node3)

	d.SetWeight(node1, 1)
	d.SetWeight(node2, 2)
	d.SetWeight(node3, 4)

	bucket1 := 0
	bucket2 := 0
	bucket3 := 0

	for range 100 {
		nodes := allNodes.Clone()
		blockId := model.NewBlockId()
		ptrs := d.PointersForId(blockId)

		if len(ptrs) != 2 {
			t.Error("should have 2 main data nodes")
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

		if ptrs[0].NodeId == node1 {
			bucket1++
		} else if ptrs[0].NodeId == node2 {
			bucket2++
		} else if ptrs[0].NodeId == node3 {
			bucket3++
		}

		if ptrs[1].NodeId == node1 {
			bucket1++
		} else if ptrs[1].NodeId == node2 {
			bucket2++
		} else if ptrs[1].NodeId == node3 {
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
