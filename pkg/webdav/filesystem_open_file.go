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
	"os"
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
	ro := os.O_RDONLY&req.flag != 0
	rw := os.O_RDWR&req.flag != 0
	wo := os.O_WRONLY&req.flag != 0
	append := os.O_APPEND&req.flag != 0
	create := os.O_CREATE&req.flag != 0
	failIfExists := os.O_EXCL&req.flag != 0
	truncate := os.O_TRUNC&req.flag != 0
	isDir := req.perm.IsDir()

	// only one of ro, rw, wo allowed
	if (ro && rw) || (ro && wo) || (rw && wo) || !(ro || rw || wo) {
		return openFileResp{err: errors.New("invalid flag")}
	}

	if ro && (append || create || failIfExists || truncate) {
		return openFileResp{err: errors.New("invalid flag")}
	}

	if !create && failIfExists {
		return openFileResp{err: errors.New("invalid flag")}
	}

	// opening the root directory
	dirName, fileName := dirAndFileName(req.name)
	if fileName == "" && dirName == "" {
		if isDir {
			file := f.FilesByPath["/"]
			return openFileResp{file: &file}
		} else {
			return openFileResp{err: errors.New("not a directory")}
		}
	}

	// handle failIfExists scenario
	file, exists := f.FilesByPath[req.name]
	if failIfExists && exists {
		return openFileResp{err: errors.New("file exists")}
	}

	// can't open a file that doesn't exist in read-only mode
	if !exists && ro {
		return openFileResp{err: errors.New("file not found")}
	}

	if exists && isDir && !file.IsDir() {
		return openFileResp{err: errors.New("file isn't directory")}
	}

	if exists && !isDir && file.IsDir() {
		return openFileResp{err: errors.New("file is directory")}
	}

	if !exists {
		file = File{
			NameValue:    fileName,
			IsDirValue:   isDir,
			RO:           ro,
			RW:           rw,
			WO:           wo,
			Append:       append,
			Create:       create,
			FailIfExists: failIfExists,
			Truncate:     truncate,
			SizeValue:    0,
			ModeValue:    0,
			Modtime:      time.Now(),
			SysValue:     nil,
			Position:     0,
			Data:         []byte{},
			BlockId:      "",
		}
		f.FilesByPath[req.name] = file
	}

	if append {
		file.Position = file.SizeValue
	}

	return openFileResp{file: &file}
}
