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

package rebalancer

import (
	"context"
	"tealfs/pkg/disk"
	"tealfs/pkg/model"
	"tealfs/pkg/set"
	"testing"
)

func TestLocalBlockIdLister(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	inFetchIds := make(chan AllBlockIdReq)
	outIdLocalResults := make(chan AllBlockIdResp)
	outIdRemoteResults := make(chan model.MgrConnsSend)
	disks := set.NewSet[disk.Disk]()
	mapper := model.NewNodeConnectionMapper()

	lister := LocalBlockIdLister{
		InFetchIds:         inFetchIds,
		OutIdLocalResults:  outIdLocalResults,
		OutIdRemoteResults: outIdRemoteResults,
		Disks:              &disks,
		NodeId:             "node1",
		Mapper:             mapper,
	}
	go lister.Start(ctx)

	inFetchIds <- AllBlockIdReq{
		Caller: "node1",
		Id:     "id1",
	}

	outLocal := <-outIdLocalResults
	if outLocal.Caller != "node1" {
		t.Errorf("unexpected caller in local response: got %s, want %s", outLocal.Caller, "node1")
	}
	if outLocal.Id != "id1" {
		t.Errorf("unexpected ID in local response: got %s, want %s", outLocal.Id, "id1")
	}
	if outLocal.BlockIds.Len() != 0 {
		t.Errorf("unexpected number of block IDs in local response: got %d, want %d", outLocal.BlockIds.Len(), 0)
	}

}
