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

package disk

import (
	"context"
	"encoding/json"
	"path/filepath"
	"tealfs/pkg/model"
	"tealfs/pkg/set"
	"testing"
)

func TestDiskSaver(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	fileOps := MockFileOps{DataWritten: make(chan struct{})}
	allDiskIds := set.NewSet[model.AddDiskReq]()
	save := make(chan struct{})

	saver := DiskSaver{
		FileOps:    &fileOps,
		LoadPath:   "loadPath",
		AllDiskIds: &allDiskIds,
		Save:       save,
	}
	go saver.Start(ctx)

	allDiskIds.Add(model.AddDiskReq{
		DiskId: "diskId1",
		Path:   "path1",
		NodeId: "nodeId1",
	})
	allDiskIds.Add(model.AddDiskReq{
		DiskId: "diskId2",
		Path:   "path2",
		NodeId: "nodeId2",
	})

	save <- struct{}{}
	<-fileOps.DataWritten

	bytes, err := fileOps.ReadFile(filepath.Join(saver.LoadPath, "disks.json"))
	if err != nil {
		t.Errorf("Didn't write a file: %v", err)
		return
	}

	loadedDisks := []model.AddDiskReq{}
	err = json.Unmarshal(bytes, &loadedDisks)
	if err != nil {
		t.Errorf("unable to parse file: %v", err)
		return
	}
	loadedDisksSet := set.NewSetFromSlice(loadedDisks)

	if !allDiskIds.Equal(&loadedDisksSet) {
		t.Error("unexpected disks")
		return
	}
}
