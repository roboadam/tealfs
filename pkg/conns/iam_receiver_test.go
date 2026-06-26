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

package conns

import (
	"context"
	"tealfs/pkg/model"
	"testing"
)

func TestIamReceiver(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	inIamTrigger := make(chan struct{})
	outSendSyncNodes := make(chan struct{}, 1)
	outSaveCluster := make(chan struct{}, 1)
	mapper := model.NewNodeConnectionMapper()

	iamReceiver := IamReceiver{
		InIamTrigger:     inIamTrigger,
		OutSendSyncNodes: outSendSyncNodes,
		OutSaveCluster:   outSaveCluster,
		Mapper:           mapper,
	}
	go iamReceiver.Start(ctx)

	inIamTrigger <- struct{}{}
	<-outSaveCluster
	<-outSendSyncNodes
}
