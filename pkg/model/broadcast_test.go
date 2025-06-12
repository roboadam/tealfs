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

package model_test

import (
	"tealfs/pkg/model"
	"testing"
)

func TestBroadcast(t *testing.T) {
	b1 := model.NewBroadcast([]byte{1, 2, 3}, model.FileSystemDest)
	b2 := model.NewBroadcast([]byte{1, 2, 3, 4}, model.FileSystemDest)
	if b1.Equal(&b2) {
		t.Errorf("b1 and b2 are not equal")
		return
	}

	raw := b1.ToBytes()
	b3 := model.ToBroadcast(raw[1:])
	if !b1.Equal(b3) {
		t.Errorf("Expected %v, got %v", b1, b3)
	}
}
