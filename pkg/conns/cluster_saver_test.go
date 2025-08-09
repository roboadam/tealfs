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

type synchronizingMockFileOps struct {
	disk.MockFileOps
	writeDone chan struct{}
}

func (m *synchronizingMockFileOps) WriteFile(path string, data []byte) error {
	err := m.MockFileOps.WriteFile(path, data)
	if m.writeDone != nil {
		m.writeDone <- struct{}{}
	}
	return err
}

func TestClusterSaver(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	save := make(chan struct{})
	nodeConnMapper := model.NewNodeConnectionMapper()
	savePath := ""
	fileOps := &synchronizingMockFileOps{
		writeDone: make(chan struct{}, 1),
	}

	clusterSaver := ClusterSaver{
		Save:           save,
		NodeConnMapper: nodeConnMapper,
		SavePath:       savePath,
		FileOps:        fileOps,
	}
	go clusterSaver.Start(ctx)

	save <- struct{}{}
	<-fileOps.writeDone

	bytes, err := fileOps.ReadFile(filepath.Join(savePath, "cluster.json"))
	if err != nil {
		t.Errorf("error reading cluster.json %v", err)
		return
	}

	newMapper, err := model.NodeConnectionMapperUnmarshal(bytes)
	if err != nil {
		t.Errorf("error unmarshaling cluster.json %v", err)
		return
	}

	newNodes := newMapper.Nodes()
	if newNodes.Len() != 0 {
		t.Errorf("expected 0 nodes, got %d", newNodes.Len())
		return
	}
}
