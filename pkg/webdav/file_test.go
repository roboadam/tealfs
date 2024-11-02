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

package webdav_test

import (
	"io"
	"tealfs/pkg/model"
	"tealfs/pkg/webdav"
	"testing"
	"time"
)

func TestSeek(t *testing.T) {
	file := webdav.File{
		IsDirValue:   false,
		RO:           false,
		RW:           false,
		WO:           false,
		Append:       false,
		Create:       false,
		FailIfExists: false,
		Truncate:     false,
		SizeValue:    0,
		ModeValue:    0,
		Modtime:      time.Now(),
		SysValue:     nil,
		Position:     0,
		Block: model.Block{
			Id:   "",
			Data: []byte{1, 2, 3, 4, 5},
		},
	}

	result, err := file.Seek(3, io.SeekStart)
	if err != nil {
		t.Error("error seeking", err)
	}
	if result != 3 {
		t.Error("position should be 3 instead of", result)
	}

	result, err = file.Seek(3, io.SeekStart)
	if err != nil {
		t.Error("error seeking", err)
	}
	if result != 3 {
		t.Error("second position should be 3 instead of", result)
	}

	result, err = file.Seek(3, io.SeekCurrent)
	if err != nil {
		t.Error("error seeking", err)
	}
	if result != 6 {
		t.Error("second position should be 6 instead of", result)
	}

	result, err = file.Seek(-4, io.SeekEnd)
	if err != nil {
		t.Error("error seeking", err)
	}
	if result != 1 {
		t.Error("second position should be 1 instead of", result)
	}
	_, err = file.Seek(-4, io.SeekCurrent)
	if err == nil {
		t.Error("position shouldn't be allowed to be negative")
	}
}
