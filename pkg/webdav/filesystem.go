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
	"tealfs/pkg/chanutil"
	"tealfs/pkg/disk"
	"tealfs/pkg/model"
	"tealfs/pkg/set"
	"time"

	log "github.com/sirupsen/logrus"
)

type FileSystem struct {
	fileHolder      FileHolder
	mkdirReq        chan mkdirReq
	openFileReq     chan openFileReq
	removeAllReq    chan removeAllReq
	renameReq       chan renameReq
	writeReq        chan writeReq
	readReq         chan readReq
	seekReq         chan seekReq
	closeReq        chan closeReq
	readdirReq      chan readdirReq
	statReq         chan statReq
	nameReq         chan nameReq
	sizeReq         chan sizeReq
	modeReq         chan modeReq
	modtimeReq      chan modtimeReq
	isdirReq        chan isdirReq
	sysReq          chan sysReq
	listBlockIdsReq chan listBlockIdsReq
	ReadReqResp     chan ReadReqResp
	WriteReqResp    chan WriteReqResp
	inBroadcast     chan FileBroadcast
	OutSends        chan model.MgrConnsSend

	Mapper    *model.NodeConnectionMapper
	nodeId    model.NodeId
	fileOps   disk.FileOps
	indexPath string
	Ctx       context.Context
}

func NewFileSystem(
	nodeId model.NodeId,
	inBroadcast chan FileBroadcast,
	fileOps disk.FileOps,
	indexPath string,
	chansize int,
	outSends chan model.MgrConnsSend,
	mapper *model.NodeConnectionMapper,
	ctx context.Context,
) FileSystem {
	filesystem := FileSystem{
		fileHolder:      NewFileHolder(),
		mkdirReq:        make(chan mkdirReq, chansize),
		openFileReq:     make(chan openFileReq, chansize),
		removeAllReq:    make(chan removeAllReq, chansize),
		renameReq:       make(chan renameReq, chansize),
		writeReq:        make(chan writeReq, chansize),
		readReq:         make(chan readReq, chansize),
		seekReq:         make(chan seekReq, chansize),
		closeReq:        make(chan closeReq, chansize),
		readdirReq:      make(chan readdirReq, chansize),
		statReq:         make(chan statReq, chansize),
		nameReq:         make(chan nameReq, chansize),
		sizeReq:         make(chan sizeReq, chansize),
		modeReq:         make(chan modeReq, chansize),
		modtimeReq:      make(chan modtimeReq, chansize),
		isdirReq:        make(chan isdirReq, chansize),
		sysReq:          make(chan sysReq, chansize),
		listBlockIdsReq: make(chan listBlockIdsReq, chansize),
		ReadReqResp:     make(chan ReadReqResp, chansize),
		WriteReqResp:    make(chan WriteReqResp, chansize),
		inBroadcast:     inBroadcast,
		nodeId:          nodeId,
		fileOps:         fileOps,
		indexPath:       indexPath,
		OutSends:        outSends,
		Mapper:          mapper,
		Ctx:             ctx,
	}
	block := model.Block{Id: model.NewBlockId(), Data: []byte{}}
	root := File{
		SizeValue:  0,
		ModeValue:  fs.ModeDir,
		Modtime:    time.Time{},
		Position:   0,
		Block:      []model.Block{block},
		HasData:    []bool{false},
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
	chanutil.Send(f.Ctx, f.ReadReqResp, ReadReqResp{req, resp}, "filesystem fetchBlock "+string(req.Id))
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
		case <-f.Ctx.Done():
			return
		case req := <-f.mkdirReq:
			chanutil.Send(f.Ctx, req.respChan, f.mkdir(&req), "filesystem: run mkdirReq")
		case req := <-f.openFileReq:
			chanutil.Send(f.Ctx, req.respChan, f.openFile(&req), "filesystem: run openFile")
		case req := <-f.removeAllReq:
			chanutil.Send(f.Ctx, req.respChan, f.removeAll(&req), "filesystem: run removeAll")
		case req := <-f.renameReq:
			chanutil.Send(f.Ctx, req.respChan, f.rename(&req), "filesystem: run rename")
		case req := <-f.writeReq:
			chanutil.Send(f.Ctx, req.resp, write(req), "filesystem: write")
		case req := <-f.readReq:
			chanutil.Send(f.Ctx, req.resp, read(req), "filesystem: read")
		case req := <-f.seekReq:
			chanutil.Send(f.Ctx, req.resp, seek(req), "filesystem: seek")
		case req := <-f.closeReq:
			chanutil.Send(f.Ctx, req.resp, closeF(req), "filesystem: close")
		case req := <-f.readdirReq:
			chanutil.Send(f.Ctx, req.resp, readdir(req), "filesystem: readdir")
		case req := <-f.statReq:
			chanutil.Send(f.Ctx, req.resp, stat(req), "filesystem: stat")
		case req := <-f.nameReq:
			chanutil.Send(f.Ctx, req.resp, name(req), "filesystem: name")
		case req := <-f.sizeReq:
			chanutil.Send(f.Ctx, req.resp, size(req), "filesystem: size")
		case req := <-f.modeReq:
			chanutil.Send(f.Ctx, req.resp, mode(req), "filesystem: mode")
		case req := <-f.modtimeReq:
			chanutil.Send(f.Ctx, req.resp, modtime(req), "filesystem: modtime")
		case req := <-f.isdirReq:
			chanutil.Send(f.Ctx, req.resp, isdir(req), "filesystem: isdir")
		case req := <-f.sysReq:
			chanutil.Send(f.Ctx, req.resp, sys(req), "filesystem: sys")
		case req := <-f.listBlockIdsReq:
			f.listBlockIds(&req)
		case msg := <-f.inBroadcast:
			file, _, err := FileFromBytes(msg.FileBytes, f)
			if err != nil {
				log.Error("Error decoding file update")
			} else {
				switch msg.UpdateType {
				case UpsertFile:
					f.fileHolder.Upsert(&file)
				case DeleteFile:
					f.fileHolder.Delete(&file)
				}
				err = f.persistFileIndex()
				if err != nil {
					log.Error("Unable to persist file index:", err)
				}
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
	if err != nil && !errors.Is(err, fs.ErrNotExist) {
		return err
	}
	return f.fileHolder.UpdateFileHolderFromBytes(data, f)
}

func (f *FileSystem) persistFileIndexAndBroadcast(file *File, updateType FileBroadcastType) error {
	err := f.persistFileIndex()
	if err != nil {
		log.Error("Error persisting index", err)
		return err
	}
	msg := FileBroadcast{UpdateType: updateType, FileBytes: file.ToBytes()}
	conns := f.Mapper.Connections()
	for _, connId := range conns.GetValues() {
		f.OutSends <- model.MgrConnsSend{ConnId: connId, Payload: &msg}
	}
	return nil
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
		Block:      []model.Block{block},
		HasData:    []bool{false},
		Path:       p,
		FileSystem: f,
	}

	f.fileHolder.Add(&dir)
	err = f.persistFileIndexAndBroadcast(&dir, UpsertFile)
	if err != nil {
		return err
	}

	return nil
}

type listBlockIdsReq struct {
	respChan chan set.Set[model.BlockId]
}

func (f *FileSystem) ListBlockIds() *set.Set[model.BlockId] {
	req := listBlockIdsReq{respChan: make(chan set.Set[model.BlockId], 1)}
	f.listBlockIdsReq <- req
	resp := <-req.respChan
	return &resp
}

func (f *FileSystem) listBlockIds(req *listBlockIdsReq) {
	result := set.NewSet[model.BlockId]()
	for blockId := range f.fileHolder.byBlockId {
		result.Add(blockId)
	}
	req.respChan <- result
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
			err = f.persistFileIndexAndBroadcast(file, DeleteFile)
			if err != nil {
				return err
			}
		}
	}
	f.fileHolder.Delete(baseFile)
	f.persistFileIndexAndBroadcast(baseFile, DeleteFile)

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
	req := renameReq{
		ctx:      ctx,
		oldName:  oldName,
		newName:  newName,
		respChan: respChan,
	}
	chanutil.Send(f.Ctx, f.renameReq, req, "rename")
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

	if file.ModeValue.IsDir() {
		for _, child := range f.fileHolder.AllFiles() {
			if child.Path.startsWith(oldPath) {
				f.fileHolder.Delete(child)
				child.Path = child.Path.swapPrefix(oldPath, newPath)
				f.fileHolder.Add(child)
				f.persistFileIndexAndBroadcast(child, UpsertFile)
			}
		}
	} else {
		f.fileHolder.Delete(file)
		file.Path = newPath
		f.fileHolder.Add(file)
		f.persistFileIndexAndBroadcast(file, UpsertFile)
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
