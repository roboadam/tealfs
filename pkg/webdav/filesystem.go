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
)

type FileSystem struct {
	FilesByPath  fileHolder
	mkdirReq     chan mkdirReq
	openFileReq  chan openFileReq
	removeAllReq chan removeAllReq
	renameReq    chan renameReq
	ReadReqResp  chan ReadReqResp
	WriteReqResp chan WriteReqResp
	nodeId       model.NodeId
}

func NewFileSystem(nodeId model.NodeId) FileSystem {
	filesystem := FileSystem{
		FilesByPath:  fileHolder{data: make(map[pathValue]File)},
		mkdirReq:     make(chan mkdirReq),
		openFileReq:  make(chan openFileReq),
		removeAllReq: make(chan removeAllReq),
		renameReq:    make(chan renameReq),
		ReadReqResp:  make(chan ReadReqResp),
		WriteReqResp: make(chan WriteReqResp),
		nodeId:       nodeId,
	}
	block := model.Block{Id: "", Data: []byte{}}
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
		Block:        block,
		hasData:      false,
		path:         []pathSeg{},
		fileSystem:   &filesystem,
	}
	filesystem.FilesByPath.add(root)
	go filesystem.run()
	return filesystem
}

type WriteReqResp struct {
	Req  model.WriteRequest
	Resp chan model.WriteResult
}

type ReadReqResp struct {
	Req  model.ReadRequest
	Resp chan model.ReadResult
}

func (f *FileSystem) fetchBlock(id model.BlockId) model.ReadResult {
	req := model.ReadRequest{
		Caller:  f.nodeId,
		BlockId: id,
	}
	resp := make(chan model.ReadResult)
	f.ReadReqResp <- ReadReqResp{req, resp}
	return <-resp
}

func (f *FileSystem) immediateChildren(path Path) []File {
	children := make([]File, 0)
	neededPathLen := len(path) + 1
	for _, file := range f.FilesByPath.allFiles() {
		if len(file.path) == neededPathLen && file.path.startsWith(path) {
			children = append(children, file)
		}
	}
	return children
}

func (f *FileSystem) pushBlock(block model.Block) model.WriteResult {
	req := model.WriteRequest{Caller: f.nodeId, Block: block}
	resp := make(chan model.WriteResult)
	f.WriteReqResp <- WriteReqResp{req, resp}
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
	p, err := PathFromName(req.name)
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

	block := model.Block{
		Id:   "",
		Data: []byte{},
	}
	dir := File{
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
		Block:        block,
		hasData:      false,
		path:         p,
		fileSystem:   f,
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
	pathToDelete, err := PathFromName(req.name)
	if err != nil {
		return err
	}
	if !f.FilesByPath.exists(pathToDelete) {
		return errors.New("file does not exist")
	}
	for _, file := range f.FilesByPath.allFiles() {
		if file.path.startsWith(pathToDelete) {
			f.FilesByPath.delete(file.path)
		}
	}
	f.FilesByPath.delete(pathToDelete)
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

func (f *FileSystem) rename(req *renameReq) error {
	oldPath, err := PathFromName(req.oldName)
	if err != nil {
		return err
	}
	newPath, err := PathFromName(req.newName)
	if err != nil {
		return err
	}

	file, exists := f.FilesByPath.get(oldPath)
	if !exists {
		return errors.New("file not found")
	}

	if file.IsDir() {
		for _, child := range f.FilesByPath.allFiles() {
			if child.path.startsWith(oldPath) {
				f.FilesByPath.delete(child.path)
				child.path = child.path.swapPrefix(oldPath, newPath)
				f.FilesByPath.add(child)
			}
		}
	} else {
		f.FilesByPath.delete(oldPath)
		file.path = newPath
		f.FilesByPath.add(file)
	}

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
		forStat:  true,
	}
	resp := <-respChan
	return resp.file, resp.err
}
