package dist_test

import (
	"tealfs/pkg/disk/dist"
	"tealfs/pkg/model"
	"tealfs/pkg/set"
	"testing"

	"github.com/google/uuid"
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
		block1 := model.BlockKeyId(uuid.NewString())
		block2 := model.BlockKeyId(uuid.NewString())
		key1, key2, err := d.KeysForIds(block1, block2)
		if err != nil {
			t.Error(err)
			return
		}

		if len(key1.Data) != 1 && len(key2.Data) != 1 {
			t.Error("should have 1 main data nodes")
			return
		}

		if !nodes.Exists(key1.Data[0].NodeId) {
			t.Error("missing one of the nodes")
			return
		}
		nodes.Remove(key1.Data[0].NodeId)

		if !nodes.Exists(key2.Data[0].NodeId) {
			t.Error("missing one of the nodes")
			return
		}
		nodes.Remove(key2.Data[0].NodeId)

		if key1.Parity.NodeId != key2.Parity.NodeId {
			t.Error("parity nodes should be the same")
			return
		}

		if !nodes.Exists(key2.Parity.NodeId) {
			t.Error("missing one of the nodes")
			return
		}

		if key1.Data[0].NodeId == node1 {
			bucket1++
		} else if key1.Data[0].NodeId == node2 {
			bucket2++
		} else if key1.Data[0].NodeId == node3 {
			bucket3++
		} else if key1.Data[0].NodeId == node4 {
			bucket4++
		}

		if key2.Data[0].NodeId == node1 {
			bucket1++
		} else if key2.Data[0].NodeId == node2 {
			bucket2++
		} else if key2.Data[0].NodeId == node3 {
			bucket3++
		} else if key2.Data[0].NodeId == node4 {
			bucket4++
		}

		if key2.Parity.NodeId == node1 {
			bucket1++
		} else if key2.Parity.NodeId == node2 {
			bucket2++
		} else if key2.Parity.NodeId == node3 {
			bucket3++
		} else if key2.Parity.NodeId == node4 {
			bucket4++
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
