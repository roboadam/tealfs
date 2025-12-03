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

package set

import (
	"testing"
)

func TestBimap(t *testing.T) {
	subject := NewBimap[string, int]()
	subject.Add("a", 1)
	subject.Add("b", 1)
	subject.Add("c", 2)
	subject.Add("c", 3)

	if len(subject.dataKj) != 2 {
		t.Errorf("invalid amount of internal data, %v, %v", subject.dataJk, subject.dataKj)
	}
}
