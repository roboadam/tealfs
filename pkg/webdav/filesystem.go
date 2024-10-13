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
)

type FileSystem struct {
	FilesByPath   fileHolder
	mkdirReq      chan mkdirReq
	openFileReq   chan openFileReq
	removeAllReq  chan removeAllReq
	renameReq     chan renameReq
	FetchBlockReq chan FetchBlockReq
}

func NewFileSystem() FileSystem {
	filesystem := FileSystem{
		FilesByPath:   fileHolder{data: make(map[pathValue]File)},
		mkdirReq:      make(chan mkdirReq),
		openFileReq:   make(chan openFileReq),
		removeAllReq:  make(chan removeAllReq),
		renameReq:     make(chan renameReq),
		FetchBlockReq: make(chan FetchBlockReq),
	}
	root := File{
		IsDirValue:   true,
		RO:           false,
		RW:           false,
		WO:           false,
		Append:       false,
		Create:       false,
		FailIfExists: false,
		Truncate:     false,
		SizeValue:    0,
		ModeValue:    fs.ModeDir,
		Modtime:      time.Time{},
		SysValue:     nil,
		Position:     0,
		Data:         []byte{},
		BlockId:      "",
		hasData:      false,
		fileSystem:   &filesystem,
		path:         []pathSeg{},
	}
	filesystem.FilesByPath.add(root)
	go filesystem.run()
	return filesystem
}

type FetchBlockReq struct {
	Id   model.BlockId
	Resp chan []byte
}

func (f *FileSystem) fetchBlock(id model.BlockId) []byte {
	resp := make(chan []byte)
	f.FetchBlockReq <- FetchBlockReq{id, resp}
	return <-resp
}

func (f *FileSystem) run() {
	for {
		select {
		case req := <-f.mkdirReq:
			req.respChan <- f.mkdir(&req)
		case req := <-f.openFileReq:
			req.respChan <- f.openFile(&req)
		case req := <-f.removeAllReq:
			req.respChan <- f.removeAll(&req)
		case req := <-f.renameReq:
			req.respChan <- f.rename(&req)
		}
	}
}

type mkdirReq struct {
	ctx      context.Context
	name     string
	perm     os.FileMode
	respChan chan error
}

func (f *FileSystem) Mkdir(ctx context.Context, name string, perm os.FileMode) error {
	respChan := make(chan error)
	f.mkdirReq <- mkdirReq{
		ctx:      ctx,
		name:     name,
		perm:     perm,
		respChan: respChan,
	}
	return <-respChan
}

func (f *FileSystem) mkdir(req *mkdirReq) error {
	p, err := pathFromName(req.name)
	if err != nil {
		return err
	}
	exists := f.FilesByPath.exists(p)
	if exists {
		return errors.New("path exists")
	}

	base, err := p.base()
	if err != nil {
		return err
	}
	exists = f.FilesByPath.exists(base)
	if !exists {
		return errors.New("invalid path")
	}

	dir := File{
		path:         p,
		IsDirValue:   true,
		RO:           false,
		RW:           false,
		WO:           false,
		Append:       false,
		Create:       false,
		FailIfExists: exists,
		Truncate:     false,
		SizeValue:    0,
		ModeValue:    0,
		Modtime:      time.Now(),
		SysValue:     nil,
		Position:     0,
		Data:         []byte{},
		BlockId:      "",
	}

	f.FilesByPath.add(dir)

	return nil
}

type removeAllReq struct {
	ctx      context.Context
	name     string
	respChan chan error
}

func (f *FileSystem) removeAll(req *removeAllReq) error {
	fileName, err := pathFromName(req.name)
	if err != nil {
		return err
	}
	prefix := fileName + "/"
	for key := range f.FilesByPath {
		if strings.HasPrefix(key, prefix) {
			delete(f.FilesByPath, key)
		}
	}
	delete(f.FilesByPath, fileName)
	return nil
}

func (f *FileSystem) RemoveAll(ctx context.Context, name string) error {
	respChan := make(chan error)
	f.removeAllReq <- removeAllReq{
		ctx:      ctx,
		name:     name,
		respChan: respChan,
	}
	return <-respChan
}

type renameReq struct {
	ctx      context.Context
	oldName  string
	newName  string
	respChan chan error
}

func (f *FileSystem) Rename(ctx context.Context, oldName string, newName string) error {
	respChan := make(chan error)
	f.renameReq <- renameReq{
		ctx:      ctx,
		oldName:  oldName,
		newName:  newName,
		respChan: respChan,
	}
	return <-respChan

}

func swapPrefix(oldPrefix string, newPrefix string, value string) string {
	return newPrefix + value[len(oldPrefix):]
}

func shortName(path string) string {
	pathArray := strings.Split(path, "/")
	if len(pathArray) <= 1 {
		return ""
	}
	return pathArray[len(pathArray)-1]
}

func (f *FileSystem) rename(req *renameReq) error {
	oldName := req.oldName
	newName := req.newName
	file, exists := f.FilesByPath[oldName]
	if !exists {
		return errors.New("file not found")
	}
	oldPrefix := oldName + "/"
	newPrefix := newName + "/"

	if file.IsDir() {
		for key := range f.FilesByPath {
			if strings.HasPrefix(key, oldPrefix) {
				childFile := f.FilesByPath[key]
				delete(f.FilesByPath, key)
				childFile.NameValue = shortName(key)
				f.FilesByPath[swapPrefix(oldPrefix, newPrefix, key)] = childFile
			}
		}
	}

	delete(f.FilesByPath, oldName)
	file.NameValue = shortName(newName)
	f.FilesByPath[newName] = file

	return nil
}

func (f *FileSystem) Stat(ctx context.Context, name string) (os.FileInfo, error) {
	respChan := make(chan openFileResp)
	f.openFileReq <- openFileReq{
		ctx:      ctx,
		name:     name,
		flag:     os.O_RDONLY,
		perm:     os.ModeExclusive,
		respChan: respChan,
	}
	resp := <-respChan
	return resp.file, resp.err
}
