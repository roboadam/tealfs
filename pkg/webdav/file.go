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

package webdav

import (
	"errors"
	"io"
	"io/fs"
	"tealfs/pkg/model"
	"time"
)

type File struct {
	IsDirValue   bool
	RO           bool
	RW           bool
	WO           bool
	Append       bool
	Create       bool
	FailIfExists bool
	Truncate     bool
	SizeValue    int64
	ModeValue    fs.FileMode
	Modtime      time.Time
	SysValue     any
	Position     int64
	Block        model.Block
	hasData      bool
	path         Path
	fileSystem   *FileSystem
}

func (f *File) Close() error {
	return nil
}

func (f *File) Read(p []byte) (n int, err error) {
	if !f.hasData {
		resp := f.fileSystem.fetchBlock(f.Block.Id)
		if resp.Ok {
			f.Block = resp.Block
			f.hasData = true
		} else {
			return 0, errors.New(resp.Message)
		}
	}

	if f.Position >= int64(len(f.Block.Data)) {
		return 0, io.EOF
	}

	start := f.Position
	end := f.Position + int64(len(p))
	if end > int64(len(f.Block.Data)) {
		end = int64(len(f.Block.Data))
	}

	copy(p, f.Block.Data[start:end])
	return int(end - start), nil
}

func (f *File) Seek(offset int64, whence int) (int64, error) {
	panic("not implemented") // TODO: Implement
}

func (f *File) Readdir(count int) ([]fs.FileInfo, error) {
	panic("not implemented") // TODO: Implement
}

func (f *File) Stat() (fs.FileInfo, error) {
	return f, nil
}

func (f *File) Write(p []byte) (n int, err error) {
	if int(f.Position)+len(p) > len(f.Block.Data) {
		newData := make([]byte, int(f.Position)+len(p))
		copy(newData, f.Block.Data)
		for _, b := range p {
			newData[f.Position] = b
			f.Position++
		}
	}
	result := f.fileSystem.pushBlock(f.Block)
	if result.Ok {
		return len(p), nil
	}
	return 0, errors.New(result.Message)
}

func (f *File) Name() string {
	name, err := f.path.head()
	if err != nil {
		return ""
	}
	return string(name)
}

func (f *File) Size() int64 {
	return f.SizeValue
}

func (f *File) Mode() fs.FileMode {
	return f.ModeValue
}

func (f *File) ModTime() time.Time {
	return f.Modtime
}

func (f *File) IsDir() bool {
	return f.IsDirValue
}

func (f *File) Sys() any {
	return f.SysValue
}
