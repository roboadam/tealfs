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
	"bytes"
	"tealfs/pkg/model"
	"testing"
)

func TestInt64(t *testing.T) {
	var expectedValue int64 = 123
	expectedRemainder := []byte{1, 2, 3}
	bytesResult := model.Int64ToBytes(expectedValue)
	bytesResult = append(bytesResult, expectedRemainder...)
	result, remainder := model.Int64FromBytes(bytesResult)

	if result != expectedValue {
		t.Error("Unexpected result")
		return
	}

	if !bytes.Equal(remainder, expectedRemainder) {
		t.Error("Unexpected remainder")
		return
	}
}
