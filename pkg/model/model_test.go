// Copyright (C) 2024 Adam Hess
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

package model_test

import (
	"tealfs/pkg/model"
	"testing"
)

func TestSyncNodes(t *testing.T) {
	n1 := struct {
		Node    model.NodeId
		Address string
	}{
		Node:    model.NewNodeId(),
		Address: "node:1",
	}
	n2 := struct {
		Node    model.NodeId
		Address string
	}{
		Node:    model.NewNodeId(),
		Address: "node:2",
	}
	sn1 := model.NewSyncNodes()
	sn1.Nodes.Add(n1)
	sn1.Nodes.Add(n2)
	sn2 := model.NewSyncNodes()
	sn2.Nodes.Add(n2)
	sn2.Nodes.Add(n1)

	if !sn1.Equal(&sn2) {
		t.Error("should be equal")
	}

	bytes1 := sn1.ToBytes()
	sn3 := model.ToSyncNodes(bytes1[1:])

	if !sn1.Equal(sn3) {
		t.Error("should be equal")
	}
}
