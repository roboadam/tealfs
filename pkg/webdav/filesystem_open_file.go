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
	"fmt"
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
	fmt.Println("OPEN")
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
	fmt.Println("openfile start")
	rw := (os.O_RDWR & req.flag) != 0
	wo := (os.O_WRONLY & req.flag) != 0
	ro := false
	if rw && wo {
		fmt.Println("openfile end 1")
		return openFileResp{err: errors.New("invalid flag")}
	}
	if !(rw || wo) {
		ro = true
	}
	append := os.O_APPEND&req.flag != 0
	create := os.O_CREATE&req.flag != 0
	failIfExists := os.O_EXCL&req.flag != 0
	truncate := os.O_TRUNC&req.flag != 0
	isDirForCreate := req.perm.IsDir()

	if ro && (append || create || failIfExists || truncate) {
		fmt.Println("openfile end 2")
		return openFileResp{err: errors.New("invalid flag")}
	}

	if !create && failIfExists {
		fmt.Println("openfile end 3")
		return openFileResp{err: errors.New("invalid flag")}
	}

	path, err := PathFromName(req.name)
	if err != nil {
		fmt.Println("openfile end 4")
		return openFileResp{err: err}
	}
	file, exists := f.FilesByPath.get(path)

	// opening the root directory
	if len(path) == 0 {
		if !exists {
			fmt.Println("openfile end 5")
			return openFileResp{err: fs.ErrNotExist}
		}
		fmt.Println("openfile end 6")
		return openFileResp{file: file}
	}

	// handle failIfExists scenario
	if failIfExists && exists {
		fmt.Println("openfile end 7")
		return openFileResp{err: fs.ErrExist}
	}

	// can't open a file that doesn't exist in read-only mode
	if !exists && ro {
		fmt.Println("openfile end 8")
		return openFileResp{err: fs.ErrNotExist}
	}

	if !exists {
		block := model.Block{Id: model.NewBlockId(), Data: []byte{}}
		file = &File{
			IsDirValue:   isDirForCreate,
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
			Block:        block,
			hasData:      false,
			path:         path,
			fileSystem:   f,
		}
		f.FilesByPath.add(file)
	}

	if append {
		file.Position = file.SizeValue
	}
	fmt.Println("openfile end 9")
	return openFileResp{file: file}
}
