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
	"io/fs"
	"os"
	"time"

	"golang.org/x/net/webdav"
)

type FileSystem struct{
	
}

func (f *FileSystem) Mkdir(ctx context.Context, name string, perm os.FileMode) error {
	panic("not implemented") // TODO: Implement
}

func (f *FileSystem) OpenFile(ctx context.Context, name string, flag int, perm os.FileMode) (webdav.File, error) {
	panic("not implemented") // TODO: Implement
}

func (f *FileSystem) RemoveAll(ctx context.Context, name string) error {
	panic("not implemented") // TODO: Implement
}

func (f *FileSystem) Rename(ctx context.Context, oldName string, newName string) error {
	panic("not implemented") // TODO: Implement
}

func (f *FileSystem) Stat(ctx context.Context, name string) (os.FileInfo, error) {
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	default:
		info := FileInfo{
			name:    name,
			size:    0,
			mode:    0,
			modtime: time.Time{},
			isdir:   false,
			sys:     nil,
		}
		return &info, nil
	}
}

type File struct{}

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
	panic("not implemented") // TODO: Implement
}

func (f *File) Write(p []byte) (n int, err error) {
	panic("not implemented") // TODO: Implement
}

type FileInfo struct {
	name    string
	size    int64
	mode    fs.FileMode
	modtime time.Time
	isdir   bool
	sys     any
}

func (f *FileInfo) Name() string {
	return f.name
}

func (f *FileInfo) Size() int64 {
	return f.size
}

func (f *FileInfo) Mode() fs.FileMode {
	return f.mode
}

func (f *FileInfo) ModTime() time.Time {
	return f.modtime
}

func (f *FileInfo) IsDir() bool {
	return f.isdir
}

func (f *FileInfo) Sys() any {
	return f.sys
}
