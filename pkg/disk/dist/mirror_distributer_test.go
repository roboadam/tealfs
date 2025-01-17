package dist_test

import (
	"tealfs/pkg/disk/dist"
	"tealfs/pkg/model"
	"tealfs/pkg/set"
	"testing"

	"github.com/google/uuid"
)

func TestWriteData(t *testing.T) {
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
		block := model.BlockKeyId(uuid.NewString())
		key := d.KeyForId(block)

		if len(key.Data) != 2 {
			t.Error("should have 2 main data nodes")
			return
		}

		if !nodes.Exists(key.Data[0].NodeId) {
			t.Error("missing one of the nodes")
			return
		}
		nodes.Remove(key.Data[0].NodeId)

		if !nodes.Exists(key.Data[1].NodeId) {
			t.Error("missing one of the nodes")
			return
		}

		if key.Data[0].NodeId == node1 {
			bucket1++
		} else if key.Data[0].NodeId == node2 {
			bucket2++
		} else if key.Data[0].NodeId == node3 {
			bucket3++
		}

		if key.Data[1].NodeId == node1 {
			bucket1++
		} else if key.Data[1].NodeId == node2 {
			bucket2++
		} else if key.Data[1].NodeId == node3 {
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
