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
	d.SetWeight(node3, 3)

	block := model.BlockKeyId(uuid.NewString())
	key := d.KeyForId(block)

	if len(key.Data) != 2 {
		t.Error("should have 2 main data nodes")
		return
	}

	if !allNodes.Exists(key.Data[0].NodeId) {
		t.Error("missing one of the nodes")
		return
	}
}
