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
	FilesByPath  map[string]File
	mkdirReq     chan mkdirReq
	openFileReq  chan openFileReq
	removeAllReq chan removeAllReq
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
	_, exists := f.FilesByPath[req.name]
	if exists {
		return errors.New("path exists")
	}

	dirName, fileName := dirAndFileName(req.name)
	_, exists = f.FilesByPath[dirName]
	if !exists {
		return errors.New("invalid path")
	}

	dir := File{
		NameValue:    fileName,
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
		IsOpen:       false,
		BlockId:      "",
	}

	f.FilesByPath[req.name] = dir

	return nil
}

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
			IsOpen:       false,
			BlockId:      "",
		}
		f.FilesByPath[req.name] = file
	}

	if append {
		file.Position = file.SizeValue
	}

	return openFileResp{file: &file}
}

type removeAllReq struct {
	ctx      context.Context
	name     string
	respChan chan error
}

func (f *FileSystem) removeAll(req *removeAllReq) error {
	fileName := req.name
	prefix := fileName + "/"
	for key := range f.FilesByPath {
		if strings.HasPrefix(key, fileName)
	}
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

func dirAndFileName(name string) (string, string) {
	raw := strings.Split(name, "/")
	result := make([]string, 0)
	for _, value := range raw {
		if value != "" {
			result = append(result, value)
		}
	}
	last := len(result) - 1
	if last < 0 {
		return "", ""
	}
	return strings.Join(result[:last], "/"), result[last]
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
