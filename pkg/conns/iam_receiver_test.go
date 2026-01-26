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
	"tealfs/pkg/model"
	"testing"
)

func TestIamReceiver(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	inIam := make(chan IamConnId)
	outSendSyncNodes := make(chan struct{}, 1)
	outSaveCluster := make(chan struct{}, 1)
	mapper := model.NewNodeConnectionMapper()

	iamReceiver := IamReceiver{
		InIam:            inIam,
		OutSendSyncNodes: outSendSyncNodes,
		OutSaveCluster:   outSaveCluster,
		Mapper:           mapper,
	}
	go iamReceiver.Start(ctx)

	inIam <- IamConnId{
		Iam: model.IAm{
			NodeId: "remoteNodeId1",
			Disks: []model.DiskInfo{{
				DiskId: "remoteDisk1",
				Path:   "remotePath1",
				NodeId: "remoteNodeId1",
			}, {
				DiskId: "remoteDisk2",
				Path:   "remotePath2",
				NodeId: "remoteNodeId1",
			}},
			Address: "remoteAddress1",
		},
		ConnId: 0,
	}
	<-outSaveCluster
	<-outSendSyncNodes

	connections := mapper.Connections()
	if connections.Len() != 1 {
		t.Errorf("expected 1 connection, got %d", connections.Len())
		return
	}

	inIam <- IamConnId{
		Iam: model.IAm{
			NodeId:  "remoteNodeId2",
			Disks:   []model.DiskInfo{},
			Address: "remoteAddress2",
		},
		ConnId: 1,
	}
	<-outSaveCluster
	<-outSendSyncNodes

	connections = mapper.Connections()
	if connections.Len() != 2 {
		t.Errorf("expected 2 connection, got %d", connections.Len())
		return
	}
}
