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
	"context"
	"errors"
	"io/fs"
	"os"
	"strings"
	"tealfs/pkg/model"
	"time"

	"golang.org/x/net/webdav"
)

type FileSystem struct {
	Root          File
	BlockRequest  chan model.BlockId
	BlockResponse chan []byte
}

func (f *FileSystem) Mkdir(ctx context.Context, name string, perm os.FileMode) error {
	// Todo handle the context
	names := paths(name)
	current := f.Root
	for _, dir := range names[:len(names)-1] {
		if hasDirWithName(&current, dir) {
			current = current.Chidren[dir]
		} else {
			return errors.New("invalid path")
		}
	}
	if hasChildWithName(&current, names[len(names)-1]) {
		return errors.New("invalid path")
	}
	current.Chidren[names[len(names)-1]] = File{
		NameValue:  names[len(names)-1],
		IsDirValue: true,
		Chidren:    map[string]File{},
	}
	return nil
}

func hasDirWithName(file *File, dirName string) bool {
	child, exists := file.Chidren[dirName]
	return exists && child.IsDirValue
}

func hasChildWithName(file *File, dirName string) bool {
	_, exists := file.Chidren[dirName]
	return exists
}

func (f *FileSystem) OpenFile(ctx context.Context, name string, flag int, perm os.FileMode) (webdav.File, error) {
	return f.openFile(name, flag, perm)
}

func (f *FileSystem) openFile(name string, flag int, perm os.FileMode) (*File, error) {
	ro := os.O_RDONLY&flag != 0
	rw := os.O_RDWR&flag != 0
	wo := os.O_WRONLY&flag != 0
	append := os.O_APPEND&flag != 0
	create := os.O_CREATE&flag != 0
	failIfExists := os.O_EXCL&flag != 0
	truncate := os.O_TRUNC&flag != 0
	isDir := perm.IsDir()

	if (ro && rw) || (ro && wo) || (rw && wo) || !(ro || rw || wo) {
		return nil, errors.New("invalid flag")
	}

	if ro && (append || create || failIfExists || truncate) {
		return nil, errors.New("invalid flag")
	}

	if !create && failIfExists {
		return nil, errors.New("invalid flag")
	}

	dirNames, fileName := paths(name)
	if fileName == nil {
		if isDir {
			return &f.Root, nil
		} else {
			return nil, errors.New("not a directory")
		}
	}
	current := f.Root
	for _, dirName := range dirNames {
		if !current.IsDirValue {
			return nil, errors.New("invalid path")
		}

		dir, exists := current.Chidren[dirName]
		if !exists {
			return nil, errors.New("invalid path")
		}
		current = dir
	}

	file, exists := current.Chidren[*fileName]

	if exists && failIfExists {
		return nil, errors.New("invalid path")
	}

	file

	current.RO = ro
	current.RW = rw
	current.WO = wo
	current.Append = append
	current.Create = create
	current.FailIfExists = failIfExists
	current.Truncate = truncate
	if append && !ro {
		current.Position = current.SizeValue
	} else {
		current.Position = 0
	}

	return &current, nil
}

func last(i int, arry []string) bool {
	return i == len(arry)-1
}

func (f *FileSystem) RemoveAll(ctx context.Context, name string) error {
	pathsArry := paths(name)
	parentName := strings.Join(pathsArry[:len(pathsArry)-1], "/")

	file, err := f.openFile(parentName, os.O_RDWR, os.ModeDir)
	if err != nil {
		return err
	}

	delete(file.Chidren, pathsArry[len(pathsArry)-1])

	return nil
}

func (f *FileSystem) Rename(ctx context.Context, oldName string, newName string) error {
	oldPathsArry := paths(oldName)
	oldParentName := strings.Join(oldPathsArry[:len(oldPathsArry)-1], "/")
	oldParent, err := f.openFile(oldParentName, os.O_RDWR, os.ModeDir)
	if err != nil {
		return err
	}

	oldSimpleName := oldPathsArry[len(oldPathsArry)-1]
	file, exists := oldParent.Chidren[oldSimpleName]
	if !exists {
		return errors.New("file does not exist")
	}

	newPathsArry := paths(newName)
	newParentName := strings.Join(newPathsArry[:len(newPathsArry)-1], "/")

	newParent, err := f.openFile(newParentName, os.O_RDWR, os.ModeDir)
	if err != nil {
		return err
	}

	file.NameValue = newPathsArry[len(newPathsArry)-1]
	delete(oldParent.Chidren, oldSimpleName)
	newParent.Chidren[file.NameValue] = file

	return nil
}

func paths(name string) ([]string, *string) {
	raw := strings.Split(name, "/")
	result := make([]string, 0)
	for _, value := range raw {
		if value != "" {
			result = append(result, value)
		}
	}
	last := len(result) - 1
	if last < 0 {
		return result, nil
	}
	return result[:last], &result[last]
}

func (f *FileSystem) Stat(ctx context.Context, name string) (os.FileInfo, error) {
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	default:
		// Todo. don't know what right mode is here
		return f.openFile(name, os.O_RDONLY, os.ModeExclusive)
	}
}

type File struct {
	NameValue    string
	IsDirValue   bool
	Chidren      map[string]File
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
	Data         []byte
	IsOpen       bool
	BlockId      model.BlockId
}

func (f *File) Close() error {
	return nil
}

func (f *File) Read(p []byte) (n int, err error) {
	panic("not implemented") // TODO: Implement
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
	panic("not implemented") // TODO: Implement
}

func (f *File) Name() string {
	return f.NameValue
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
