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
	"tealfs/pkg/model"
	"tealfs/pkg/set"
	"testing"
)

func TestIamSender(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	inIamDiskUpdate := make(chan struct{})
	outSends := make(chan model.MgrConnsSend)
	iamSender := IamSender{
		InIamDiskUpdate:  inIamDiskUpdate,
		OutSends:         outSends,
		Mapper:           model.NewNodeConnectionMapper(),
		NodeId:           "localNodeId",
		Address:          "localAddress",
		LocalDiskSvcList: &set.Set[Disk]{},
	}
	iamSender.Mapper.SetAll(0, "remoteAddress1", "remoteNodeId1")
	iamSender.Mapper.SetAll(1, "remoteAddress2", "remoteNodeId2")
	go iamSender.Start(ctx)

	inIamDiskUpdate <- struct{}{}
	output1 := <-outSends
	output2 := <-outSends

	select {
	case <-outSends:
		t.Error("too many sends")
	default:
	}

	if iam1, ok := output1.Payload.(*model.IAm); ok {
		if len(iam1.Disks) != 2 {
			t.Error("wrong number of disks")
		}
	} else {
		t.Error("wrong type")
	}

	if iam2, ok := output2.Payload.(*model.IAm); ok {
		if len(iam2.Disks) != 2 {
			t.Error("wrong number of disks")
		}
	} else {
		t.Error("wrong type")
	}
}
