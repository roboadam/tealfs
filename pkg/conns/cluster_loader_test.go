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

package conns

import (
	"context"
	"path/filepath"
	"tealfs/pkg/disk"
	"tealfs/pkg/model"
	"testing"
)

func TestClusterLoader(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	nodeConnMapper := model.NewNodeConnectionMapper()
	nodeConnMapper.SetAll(0, "address0", "nodeId0")
	nodeConnMapper.SetAll(1, "address1", "nodeId1")
	nodeConnMapper.SetAll(2, "address2", "nodeId2")

	fileOps := disk.MockFileOps{}
	bytes, err := nodeConnMapper.Marshal()

	if err != nil {
		t.Error("error serializing nodeConnMapper")
		return
	}

	fileOps.WriteFile(filepath.Join("savePath", "cluster.json"), bytes)

	nodeConnMapper = model.NewNodeConnectionMapper()
	clusterLoader := ClusterLoader{
		NodeConnMapper: nodeConnMapper,
		FileOps:        &fileOps,
		SavePath:       "savePath",
	}
	clusterLoader.Load(ctx)

	addresses := nodeConnMapper.AddressesWithoutConnections()
	if addresses.Len() != 3 {
		t.Errorf("expected 3 addresses, got %d", addresses.Len())
		return
	}
	nodes := nodeConnMapper.Nodes()
	if nodes.Len() != 3 {
		t.Errorf("expected 3 nodes, got %d", nodes.Len())
		return
	}
}
