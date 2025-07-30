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
	"reflect"
	"tealfs/pkg/model"
	"testing"

	"github.com/google/uuid"
)

func TestDiskLoader(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	fileOps := MockFileOps{}
	disks := []model.AddDiskReq{
		{
			DiskId: model.DiskId(uuid.NewString()),
			Path:   "path1",
			NodeId: model.NewNodeId(),
		}, {
			DiskId: model.DiskId(uuid.NewString()),
			Path:   "path2",
			NodeId: model.NewNodeId(),
		},
	}
	rawBytes, _ := json.Marshal(disks)
	fileOps.WriteFile(filepath.Join("somePath", "disks.json"), rawBytes)

	outAddDisk := make(chan model.AddDiskReq)
	dl := DiskLoader{
		FileOps:    &fileOps,
		SavePath:   "somePath",
		OutAddDisk: outAddDisk,
	}
	go dl.LoadDisks(ctx)

	loadedDisk1 := <- outAddDisk
	loadedDisk2 := <- outAddDisk
	select{
	case <- outAddDisk:
		t.Error("should be now more disks to load")
		return
	default:
	}

	if !reflect.DeepEqual(disks[0], loadedDisk1) {
		t.Error("disk one not equal")
		return
	}

	if !reflect.DeepEqual(disks[1], loadedDisk2) {
		t.Error("disk two not equal")
		return
	}
}
