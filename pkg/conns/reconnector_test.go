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

package conns_test

import (
	"context"
	"tealfs/pkg/conns"
	"tealfs/pkg/model"
	"testing"
	"time"
)

func TestReconnector(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	outConnectTo := make(chan model.ConnectToNodeReq, 10)
	mapper := model.NewNodeConnectionMapper()

	mapper.SetNodeAddress("node1", "address1")
	mapper.SetAll(model.ConnId(1), "address2", "node2")

	reconnector := &conns.Reconnector{
		OutConnectTo: outConnectTo,
		Mapper:       mapper,
	}

	go reconnector.Start(ctx)

	select {
	case req := <-outConnectTo:
		if req.Address != "address1" {
			t.Errorf("expected to connect to 'address1', got '%s'", req.Address)
		}
	case <-time.After(100 * time.Millisecond):
		t.Fatal("timed out waiting for initial connection request")
	}

	select {
	case req := <-outConnectTo:
		t.Fatalf("received unexpected connection request for '%s'", req.Address)
	default:
	}

	select {
	case req := <-outConnectTo:
		if req.Address != "address1" {
			t.Errorf("expected to reconnect to 'address1', got '%s'", req.Address)
		}
	case <-time.After(1100 * time.Millisecond):
		t.Fatal("timed out waiting for scheduled connection request")
	}
}
