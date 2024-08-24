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
	"os"

	"golang.org/x/net/webdav"
)

type FileSystem struct{}

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
	panic("not implemented") // TODO: Implement
}
