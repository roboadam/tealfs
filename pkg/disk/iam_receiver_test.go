// Copyright (C) 2026 Adam Hess
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
	"tealfs/pkg/model"
	"testing"

	"github.com/google/uuid"
)

func TestIamReceiver(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	inIam := make(chan model.IAm)

	diskAddedMsg := make(chan model.DiskAddedMsg, 1)

	receiver := IamReceiver{
		InIam:           inIam,
		OutDiskAddedMsg: diskAddedMsg,
	}
	go receiver.Start(ctx)

	nodeId := model.NewNodeId()
	disks := []model.DiskInfo{
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
	<-diskAddedMsg
	<-diskAddedMsg
}
