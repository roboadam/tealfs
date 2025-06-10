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

package mgr_test

import (
	"tealfs/pkg/disk"
	"tealfs/pkg/mgr"
	"tealfs/pkg/model"
	"tealfs/pkg/set"
	"testing"
)

func TestGlobalBlockListSaving(t *testing.T) {
	gbl := set.NewSet[model.BlockId]()
	gbl.Add("blockId1")
	gbl.Add("blockId2")
	fs := disk.MockFileOps{}
	err := mgr.SaveGBL(&fs, "bl", &gbl)
	if err != nil {
		t.Errorf("Error saving GBL %v", err)
		return
	}
	gbl2, err := mgr.LoadGBL(&fs, "bl")
	if err != nil {
		t.Errorf("Error loading GBL %v", err)
		return
	}
	if !gbl2.Equal(&gbl) {
		t.Errorf("gbl1 != gbl2 %d != %d", gbl.Len(), gbl2.Len())
		return
	}

}

func TestGlobalBlockListCommand(t *testing.T) {
	cmd := mgr.GlobalBlockListCommand{
		Type:    mgr.Delete,
		BlockId: model.NewBlockId(),
	}

	byteCmd := cmd.ToBytes()

	cmd2 := mgr.ToGlobalBlockListCommand(byteCmd)

	if cmd.BlockId != cmd2.BlockId {
		t.Errorf("Expected %s got %s", cmd.BlockId, cmd2.BlockId)
		return
	}

	if cmd.Type != cmd2.Type {
		t.Errorf("Expected %d got %d", cmd.Type, cmd2.Type)
		return
	}
}
