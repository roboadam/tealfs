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
	"fmt"
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
	fmt.Println("=== File.Close ===")
	f.Position = 0
	f.Block.Data = []byte{}
	f.hasData = false
	fmt.Println("*** File.Close ***")
	return nil
}

func (f *File) Read(p []byte) (n int, err error) {
	fmt.Println("=== File.Read ===")
	error := f.ensureData()
	if error != nil {
		fmt.Println("*** File.Read 1 ***")
		return 0, error
	}

	if f.Position >= int64(len(f.Block.Data)) {
		fmt.Println("*** File.Read 2 ***")
		return 0, io.EOF
	}

	start := f.Position
	end := f.Position + int64(len(p))
	if end > int64(len(f.Block.Data)) {
		end = int64(len(f.Block.Data))
	}

	copy(p, f.Block.Data[start:end])
	fmt.Println("*** File.Read 3 ***")
	return int(end - start), nil
}

func (f *File) Seek(offset int64, whence int) (int64, error) {
	fmt.Println("=== File.Seek ===")
	newPosition := 0
	switch whence {
	case io.SeekStart:
		newPosition = int(offset)
	case io.SeekCurrent:
		newPosition = int(f.Position) + int(offset)
	case io.SeekEnd:
		newPosition = len(f.Block.Data) + int(offset)
	}
	if newPosition < 0 {
		fmt.Println("*** File.Seek 1 ***")
		return f.Position, errors.New("negative seek")
	}
	f.Position = int64(newPosition)
	fmt.Println("*** File.Seek 2 ***")
	return f.Position, nil
}

func (f *File) Readdir(count int) ([]fs.FileInfo, error) {
	fmt.Println("=== File.Readdir ===")
	if count < 0 {
		fmt.Println("*** File.Readdir 1 ***")
		return nil, errors.New("negative dir count requested")
	}
	children := f.fileSystem.immediateChildren(f.path)[count:]
	result := make([]fs.FileInfo, 0, len(children))
	for _, child := range children {
		result = append(result, &child)
	}
	fmt.Println("*** File.Readdir 2 ***")
	return result, nil
}

func (f *File) Stat() (fs.FileInfo, error) {
	fmt.Println("=== File.Stat ===")
	fmt.Println("*** File.Stat 1 ***")
	return f, nil
}

func (f *File) Write(p []byte) (n int, err error) {
	fmt.Println("=== File.Write ===")
	error := f.ensureData()
	if error != nil {
		fmt.Println("*** File.Write 1 ***")
		return 0, error
	}

	if int(f.Position)+len(p) > len(f.Block.Data) {
		needToGrow := int(f.Position) + len(p) - len(f.Block.Data)
		newData := make([]byte, needToGrow)
		f.Block.Data = append(f.Block.Data, newData...)
		f.SizeValue = int64(len(f.Block.Data))
		for _, b := range p {
			f.Block.Data[f.Position] = b
			f.Position++
		}
	}

	fmt.Println("=== File.Write 1 ===")
	result := f.fileSystem.pushBlock(f.Block)
	fmt.Println("=== File.Write 2 ===")
	if result.Ok {
		fmt.Println("*** File.Write 2 ***")
		return len(p), nil
	}
	fmt.Println("*** File.Write 3 ***")
	return 0, errors.New(result.Message)
}

func (f *File) ensureData() error {
	if !f.hasData {
		resp := f.fileSystem.fetchBlock(f.Block.Id)
		if resp.Ok {
			f.Block = resp.Block
			f.hasData = true
		} else {
			return errors.New(resp.Message)
		}
	}
	return nil
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
