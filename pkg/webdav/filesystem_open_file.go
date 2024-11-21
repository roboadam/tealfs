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
	"tealfs/pkg/model"
	"time"

	"golang.org/x/net/webdav"
)

type openFileReq struct {
	ctx      context.Context
	name     string
	flag     int
	perm     os.FileMode
	respChan chan openFileResp
}

type openFileResp struct {
	file *File
	err  error
}

func (f *FileSystem) OpenFile(ctx context.Context, name string, flag int, perm os.FileMode) (webdav.File, error) {
	respChan := make(chan openFileResp)
	f.openFileReq <- openFileReq{
		ctx:      ctx,
		name:     name,
		flag:     flag,
		perm:     perm,
		respChan: respChan,
	}
	resp := <-respChan
	return resp.file, resp.err
}

func (f *FileSystem) openFile(req *openFileReq) openFileResp {
	rw := (os.O_RDWR & req.flag) != 0
	wo := (os.O_WRONLY & req.flag) != 0
	ro := false
	if rw && wo {
		return openFileResp{err: errors.New("invalid flag")}
	}
	if !(rw || wo) {
		ro = true
	}
	append := os.O_APPEND&req.flag != 0
	create := os.O_CREATE&req.flag != 0
	failIfExists := os.O_EXCL&req.flag != 0
	truncate := os.O_TRUNC&req.flag != 0

	if ro && (append || create || failIfExists || truncate) {
		return openFileResp{err: errors.New("invalid flag")}
	}

	if !create && failIfExists {
		return openFileResp{err: errors.New("invalid flag")}
	}

	path, err := PathFromName(req.name)
	if err != nil {
		return openFileResp{err: err}
	}
	file, exists := f.FilesByPath.get(path)

	// opening the root directory
	if len(path) == 0 {
		if !exists {
			return openFileResp{err: fs.ErrNotExist}
		}
		return openFileResp{file: file}
	}

	// handle failIfExists scenario
	if failIfExists && exists {
		return openFileResp{err: fs.ErrExist}
	}

	// can't open a file that doesn't exist in read-only mode
	if !exists && ro {
		return openFileResp{err: fs.ErrNotExist}
	}

	if !exists {
		block := model.Block{Id: model.NewBlockId(), Data: []byte{}}
		file = &File{
			SizeValue:  0,
			ModeValue:  req.perm,
			Modtime:    time.Now(),
			Position:   0,
			Block:      block,
			HasData:    false,
			Path:       path,
			FileSystem: f,
		}
		f.FilesByPath.add(file)
	}

	if append {
		file.Position = file.SizeValue
	}
	return openFileResp{file: file}
}
