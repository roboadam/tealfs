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

package model_test

import (
	"tealfs/pkg/model"
	"testing"
)

const MainNodeId model.NodeId = "main"
const NonMainNode1 model.NodeId = "non-main1"
const NonMainNode2 model.NodeId = "non-main2"

func TestMainNodeCalculation(t *testing.T) {
	mapper := model.NewNodeConnectionMapper()
	mapper.SetNodeAddress(MainNodeId, "address1")
	mapper.SetNodeAddress(NonMainNode1, "address2")
	mapper.SetNodeAddress(NonMainNode2, "address3")

	main := mapper.MainNode()
	if main != MainNodeId {
		t.Error("Unexpected main node")
	}
}
