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
	"reflect"
	"tealfs/pkg/model"
	"testing"

	"github.com/google/uuid"
)

func TestIamReceiver(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	inIam := make(chan model.IAm)
	outAddDiskReq := make(chan model.AddDiskReq)
	receiver := IamReceiver{
		InIam:         inIam,
		OutAddDiskReq: outAddDiskReq,
	}
	go receiver.Start(ctx)

	nodeId := model.NewNodeId()
	disks := []model.AddDiskReq{
		{
			DiskId: model.DiskId(uuid.NewString()),
			Path:   "path1",
			NodeId: nodeId,
		},
		{
			DiskId: model.DiskId(uuid.NewString()),
			Path:   "path2",
			NodeId: nodeId,
		},
	}
	address := "someAddress"
	inIam <- model.NewIam(nodeId, disks, address)
	disk1Received := <- outAddDiskReq
	disk2Received := <- outAddDiskReq

	select {
	case <- outAddDiskReq:
		t.Error("should be just two disk requests")
		return
	default:
	}

	if !reflect.DeepEqual(disks[0], disk1Received) {
		t.Error("invalid first disk")
		return
	}
	if !reflect.DeepEqual(disks[1], disk2Received) {
		t.Error("invalid second disk")
		return
	}
}
