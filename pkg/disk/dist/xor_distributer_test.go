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

func TestXor(t *testing.T) {
	d := dist.NewXorDistributer()
	node1 := model.NewNodeId()
	node2 := model.NewNodeId()
	node3 := model.NewNodeId()
	node4 := model.NewNodeId()
	allNodes := set.NewSet[model.NodeId]()
	allNodes.Add(node1)
	allNodes.Add(node2)
	allNodes.Add(node3)
	allNodes.Add(node4)

	d.SetWeight(node1, 1)
	d.SetWeight(node2, 2)
	d.SetWeight(node3, 4)
	d.SetWeight(node4, 8)

	bucket1 := 0
	bucket2 := 0
	bucket3 := 0
	bucket4 := 0

	for range 1000 {
		nodes := allNodes.Clone()
		blockId1 := model.NewBlockId()
		blockId2 := model.NewBlockId()
		block1 := model.Block{
			Id:   blockId1,
			Data: []byte{0x0F, 0xF0, 0x0F},
		}
		block2 := model.Block{
			Id:   blockId2,
			Data: []byte{0xF0, 0x0F, 0xF0},
		}
		rawDatas, err := d.RawDataForBlocks(block1, block2)
		if err != nil {
			t.Error(err)
			return
		}

		if len(rawDatas) != 3 {
			t.Error("should have 3 places to store data")
			return
		}

		if !nodes.Exists(rawDatas[0].Ptr.NodeId) {
			t.Error("missing one of the nodes")
			return
		}
		nodes.Remove(rawDatas[0].Ptr.NodeId)

		if !nodes.Exists(rawDatas[1].Ptr.NodeId) {
			t.Error("missing one of the nodes")
			return
		}
		nodes.Remove(rawDatas[1].Ptr.NodeId)

		if !nodes.Exists(rawDatas[2].Ptr.NodeId) {
			t.Error("missing one of the nodes")
			return
		}

		for _, rd := range rawDatas {
			if rd.Ptr.NodeId == node1 {
				bucket1++
			} else if rd.Ptr.NodeId == node2 {
				bucket2++
			} else if rd.Ptr.NodeId == node3 {
				bucket3++
			} else if rd.Ptr.NodeId == node4 {
				bucket4++
			}
		}
	}

	if bucket1+bucket2+bucket3+bucket4 != 3000 {
		t.Error("should have 3000 blocks")
		return
	}

	if !(bucket1 < bucket2 && bucket2 < bucket3 && bucket3 < bucket4) {
		t.Error("should be distributed")
		return
	}
}
