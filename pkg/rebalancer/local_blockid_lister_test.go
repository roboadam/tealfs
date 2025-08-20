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

	lister := LocalBlockIdLister{
		InFetchIds:         inFetchIds,
		OutIdLocalResults:  outIdLocalResults,
		OutIdRemoteResults: outIdRemoteResults,
		Disks:              &disks,
		NodeId:             "node1",
		Mapper:             &model.NodeConnectionMapper{},
	}
	go lister.Start(ctx)
}
