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

package webdav_test

import (
	"tealfs/pkg/model"
	"tealfs/pkg/webdav"
	"testing"
	"time"
)

func TestSerializeFileHolder(t *testing.T) {
	path1, _ := webdav.PathFromName("/hello/world")
	path2, _ := webdav.PathFromName("/hello/planet")
	file1 := webdav.File{
		SizeValue: 1,
		ModeValue: 2,
		Modtime:   time.Unix(12345, 0),
		Block:     model.Block{Id: model.NewBlockId()},
		Path:      path1,
	}
	file2 := webdav.File{
		SizeValue: 3,
		ModeValue: 4,
		Modtime:   time.Unix(67890, 0),
		Block:     model.Block{Id: model.NewBlockId()},
		Path:      path2,
	}

	fh := webdav.NewFileHolder()
	fh.Add(&file1)
	fh.Add(&file2)

	fhAsBytes := fh.ToBytes()
	fh2 := webdav.NewFileHolder()
	err := fh2.UpdateFileHolderFromBytes(fhAsBytes, &webdav.FileSystem{})

	if err != nil {
		t.Error("error deserializing fileHolder")
		return
	}

	if len(fh2.AllFiles()) != 2 {
		t.Error("wrong number of files")
		return
	}

	file1Copy, exists := fh2.Get(path1)
	if !exists {
		t.Error("file1 doesn't exist")
		return
	}

	file2Copy, exists := fh2.Get(path2)
	if !exists {
		t.Error("file2 doesn't exist")
		return
	}

	if !FileEquality(&file1, file1Copy) {
		t.Error("file1 didn't deserialize properly")
		return
	}

	if !FileEquality(&file2, file2Copy) {
		t.Error("file2 didn't deserialize properly")
		return
	}
}

func FileEquality(f1 *webdav.File, f2 *webdav.File) bool {
	if f1.Block.Id != f2.Block.Id {
		return false
	}
	if f1.Size() != f2.Size() {
		return false
	}
	if f1.Mode() != f2.Mode() {
		return false
	}
	if f1.ModTime() != f2.ModTime() {
		return false
	}
	if !f1.Path.Equals(f2.Path) {
		return false
	}
	return true
}
