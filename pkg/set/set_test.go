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

package set_test

import (
	"tealfs/pkg/set"
	"testing"
)

func TestSetPop(t *testing.T) {
	s := set.NewSet[int]()
	_, _, ok := s.Pop()

	if ok {
		t.Error("should be nothing to pop")
	}

	s.Add(1)
	s.Add(2)
	s.Add(3)
	s.Add(4)

	k, remainder, ok := s.Pop()

	if !ok {
		t.Error("should be something to pop")
	}

	if !s.Contains(k) {
		t.Error("pop value not in original set")
	}

	if remainder.Contains(k) {
		t.Error("pop value in remainder set")
	}

	if remainder.Len() != 3 {
		t.Error("should only be 3 left")
	}
}
