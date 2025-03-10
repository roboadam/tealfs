// Copyright (C) 2025 Adam Hess
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
	"path/filepath"
	"tealfs/pkg/disk"
	"tealfs/pkg/model"
	"time"

	log "github.com/sirupsen/logrus"
)

type FileSystem struct {
	fileHolder   FileHolder
	mkdirReq     chan mkdirReq
	openFileReq  chan openFileReq
	removeAllReq chan removeAllReq
	renameReq    chan renameReq
	ReadReqResp  chan ReadReqResp
	WriteReqResp chan WriteReqResp
	inBroadcast  chan model.Broadcast
	outBroadcast chan model.Broadcast
	nodeId       model.NodeId
	fileOps      disk.FileOps
	indexPath    string
}

func NewFileSystem(
	nodeId model.NodeId,
	inBroadcast chan model.Broadcast,
	outBroadcast chan model.Broadcast,
	fileOps disk.FileOps,
	indexPath string,
) FileSystem {
	filesystem := FileSystem{
		fileHolder:   NewFileHolder(),
		mkdirReq:     make(chan mkdirReq),
		openFileReq:  make(chan openFileReq),
		removeAllReq: make(chan removeAllReq),
		renameReq:    make(chan renameReq),
		ReadReqResp:  make(chan ReadReqResp),
		WriteReqResp: make(chan WriteReqResp),
		inBroadcast:  inBroadcast,
		outBroadcast: outBroadcast,
		nodeId:       nodeId,
		fileOps:      fileOps,
		indexPath:    indexPath,
	}
	block := model.Block{Id: model.NewBlockId(), Data: []byte{}}
	root := File{
		SizeValue:  0,
		ModeValue:  fs.ModeDir,
		Modtime:    time.Time{},
		Position:   0,
		Block:      block,
		HasData:    false,
		Path:       []pathSeg{},
		FileSystem: &filesystem,
	}
	filesystem.fileHolder.Add(&root)
	err := filesystem.initFileIndex()
	if err != nil {
		log.Error("Unable to read fileIndex on startup:", err)
	}
	go filesystem.run()
	return filesystem
}

type WriteReqResp struct {
	Req  model.PutBlockReq
	Resp chan model.PutBlockResp
}

type ReadReqResp struct {
	Req  model.GetBlockReq
	Resp chan model.GetBlockResp
}

func (f *FileSystem) fetchBlock(req model.GetBlockReq) model.GetBlockResp {
	resp := make(chan model.GetBlockResp)
	f.ReadReqResp <- ReadReqResp{req, resp}
	return <-resp
}

func (f *FileSystem) immediateChildren(path Path) []*File {
	children := make([]*File, 0)
	neededPathLen := len(path) + 1
	for _, file := range f.fileHolder.AllFiles() {
		if len(file.Path) == neededPathLen && file.Path.startsWith(path) {
			children = append(children, file)
		}
	}
	return children
}

func (f *FileSystem) pushBlock(req model.PutBlockReq) model.PutBlockResp {
	resp := make(chan model.PutBlockResp)
	f.WriteReqResp <- WriteReqResp{Req: req, Resp: resp}
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
		case r := <-f.inBroadcast:
			msg, err := broadcastMessageFromBytes(r.Msg(), f)
			if err == nil {
				switch msg.bType {
				case upsertFile:
					f.fileHolder.Upsert(&msg.file)
					err := f.persistFileIndex()
					if err != nil {
						log.Error("Unable to persist file index:", err)
					}
				case deleteFile:
					f.fileHolder.Delete(&msg.file)
				}
			} else {
				log.Warn("Unable to parse incoming broadcast message")
			}

		}
	}
}

type mkdirReq struct {
	ctx      context.Context
	name     string
	perm     os.FileMode
	respChan chan error
}

func (f *FileSystem) initFileIndex() error {
	data, err := f.fileOps.ReadFile(filepath.Join(f.indexPath, "fileIndex"))
	if err != nil {
		return err
	}
	return f.fileHolder.UpdateFileHolderFromBytes(data, f)
}

func (f *FileSystem) persistFileIndex() error {
	return f.fileOps.WriteFile(filepath.Join(f.indexPath, "fileIndex"), f.fileHolder.ToBytes())
}

func (f *FileSystem) Mkdir(ctx context.Context, name string, perm os.FileMode) error {
	respChan := make(chan error)
	f.mkdirReq <- mkdirReq{
		ctx:      ctx,
		name:     name,
		perm:     perm,
		respChan: respChan,
	}
	result := <-respChan
	return result
}

func (f *FileSystem) mkdir(req *mkdirReq) error {
	p, err := PathFromName(req.name)
	if err != nil {
		return err
	}

	exists := f.fileHolder.Exists(p)
	if exists {
		return errors.New("path exists")
	}

	base, err := p.base()
	if err != nil {
		return err
	}
	exists = f.fileHolder.Exists(base)
	if !exists {
		return errors.New("invalid path")
	}

	block := model.Block{
		Id:   model.NewBlockId(),
		Data: []byte{},
	}
	dir := File{
		SizeValue:  0,
		ModeValue:  fs.ModeDir,
		Modtime:    time.Now(),
		Position:   0,
		Block:      block,
		HasData:    false,
		Path:       p,
		FileSystem: f,
	}

	f.fileHolder.Add(&dir)
	msg := broadcastMessage{bType: upsertFile, file: dir}
	f.outBroadcast <- model.NewBroadcast(msg.toBytes())

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

	baseFile, exists := f.fileHolder.Get(pathToDelete)
	if !exists {
		return errors.New("file does not exist")
	}
	for _, file := range f.fileHolder.AllFiles() {
		if file.Path.startsWith(pathToDelete) {
			f.fileHolder.Delete(file)
			msg := broadcastMessage{bType: deleteFile, file: *file}
			f.outBroadcast <- model.NewBroadcast(msg.toBytes())
		}
	}
	f.fileHolder.Delete(baseFile)
	msg := broadcastMessage{bType: deleteFile, file: *baseFile}
	f.outBroadcast <- model.NewBroadcast(msg.toBytes())

	return nil
}

func (f *FileSystem) RemoveAll(ctx context.Context, name string) error {
	respChan := make(chan error)
	f.removeAllReq <- removeAllReq{
		ctx:      ctx,
		name:     name,
		respChan: respChan,
	}
	resp := <-respChan
	return resp
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
	resp := <-respChan
	return resp

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

	file, exists := f.fileHolder.Get(oldPath)
	if !exists {
		return errors.New("file not found")
	}

	if file.IsDir() {
		for _, child := range f.fileHolder.AllFiles() {
			if child.Path.startsWith(oldPath) {
				f.fileHolder.Delete(child)
				child.Path = child.Path.swapPrefix(oldPath, newPath)
				f.fileHolder.Add(child)
				msg := broadcastMessage{bType: upsertFile, file: *child}
				f.outBroadcast <- model.NewBroadcast(msg.toBytes())
			}
		}
	} else {
		f.fileHolder.Delete(file)
		file.Path = newPath
		f.fileHolder.Add(file)
		msg := broadcastMessage{bType: upsertFile, file: *file}
		f.outBroadcast <- model.NewBroadcast(msg.toBytes())
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
	}
	resp := <-respChan
	return resp.file, resp.err
}
