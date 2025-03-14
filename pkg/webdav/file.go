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
	"errors"
	"io"
	"io/fs"
	"tealfs/pkg/chanutil"
	"tealfs/pkg/model"
	"time"

	log "github.com/sirupsen/logrus"
)

type File struct {
	SizeValue  int64
	ModeValue  fs.FileMode
	Modtime    time.Time
	Position   int64
	Block      model.Block
	HasData    bool
	Path       Path
	FileSystem *FileSystem
}

func (f *File) ToBytes() []byte {
	value := model.IntToBytes(uint32(f.SizeValue))
	value = append(value, model.IntToBytes(uint32(f.ModeValue))...)
	value = append(value, model.IntToBytes(uint32(f.Modtime.Unix()))...)
	value = append(value, model.StringToBytes(string(f.Block.Id))...)
	value = append(value, model.StringToBytes(string(f.Path.toName()))...)
	return value
}

func FileFromBytes(raw []byte, fileSystem *FileSystem) (File, []byte, error) {
	size, remainder := model.IntFromBytes(raw)
	mode, remainder := model.IntFromBytes(remainder)
	modtimeRaw, remainder := model.IntFromBytes(remainder)
	blockId, remainder := model.StringFromBytes(remainder)
	rawPath, remainder := model.StringFromBytes(remainder)

	path, err := PathFromName(rawPath)
	if err != nil {
		return File{}, nil, err
	}
	modTime := time.Unix(int64(modtimeRaw), 0)
	return File{
		SizeValue: int64(size),
		ModeValue: fs.FileMode(mode),
		Modtime:   modTime,
		Position:  0,
		Block: model.Block{
			Id:   model.BlockId(blockId),
			Data: []byte{},
		},
		HasData:    false,
		Path:       path,
		FileSystem: fileSystem,
	}, remainder, nil
}

func (f *File) Close() error {
	f.Position = 0
	f.Block.Data = []byte{}
	f.HasData = false
	return nil
}

type readReq struct {
	p    []byte
	f    *File
	resp chan readResp
}
type readResp struct {
	n   int
	err error
}

func (f *File) Read(p []byte) (n int, err error) {
	req := readReq{
		p:    p,
		f:    f,
		resp: make(chan readResp),
	}
	chanutil.Send(f.FileSystem.readReq, req, "read")
	resp := <-req.resp
	return resp.n, resp.err
}
func read(req readReq) readResp {
	f := req.f
	p := req.p
	err := f.ensureData()
	if err != nil {
		log.Warn("Error reading data for ", f.Name())
		return readResp{err: err}
	}

	if f.Position >= int64(len(f.Block.Data)) {
		log.Warn("EOF reading data for ", f.Name())
		return readResp{err: io.EOF}
	}

	start := f.Position
	end := f.Position + int64(len(p))
	if end > int64(len(f.Block.Data)) {
		end = int64(len(f.Block.Data))
	}

	copy(p, f.Block.Data[start:end])
	bytesRead := int(end - start)
	f.Position += int64(bytesRead)
	return readResp{n: bytesRead}
}

func (f *File) Seek(offset int64, whence int) (int64, error) {
	newPosition := 0
	switch whence {
	case io.SeekStart:
		newPosition = int(offset)
	case io.SeekCurrent:
		newPosition = int(f.Position) + int(offset)
	case io.SeekEnd:
		newPosition = int(f.SizeValue + offset)
	}
	if newPosition < 0 {
		return f.Position, errors.New("negative seek")
	}
	f.Position = int64(newPosition)
	return f.Position, nil
}

func (f *File) Readdir(count int) ([]fs.FileInfo, error) {
	if count < 0 {
		return nil, errors.New("negative dir count requested")
	}
	children := f.FileSystem.immediateChildren(f.Path)[count:]
	result := make([]fs.FileInfo, 0, len(children))
	for _, child := range children {
		result = append(result, child)
	}
	return result, nil
}

func (f *File) Stat() (fs.FileInfo, error) {
	return f, nil
}

type writeReq struct {
	p    []byte
	f    *File
	resp chan writeResp
}
type writeResp struct {
	n   int
	err error
}

func (f *File) Write(p []byte) (n int, err error) {
	req := writeReq{
		p:    p,
		f:    f,
		resp: make(chan writeResp),
	}
	chanutil.Send(f.FileSystem.writeReq, req, "write")
	resp := <-req.resp
	return resp.n, resp.err
}

func write(wreq writeReq) writeResp {
	f := wreq.f
	p := wreq.p
	err := f.ensureData()
	if err != nil {
		return writeResp{err: err}
	}

	if int(f.Position)+len(p) > len(f.Block.Data) {
		needToGrow := int(f.Position) + len(p) - len(f.Block.Data)
		newData := make([]byte, needToGrow)
		f.Block.Data = append(f.Block.Data, newData...)
		f.SizeValue = int64(len(f.Block.Data))
		for _, b := range p {
			f.Block.Data[f.Position] = b
			f.Position++
		}
	}

	req := model.NewPutBlockReq(f.Block)
	result := f.FileSystem.pushBlock(req)
	if result.Err == nil {
		err := f.FileSystem.persistFileIndexAndBroadcast(f, upsertFile)
		if err != nil {
			return writeResp{n: len(p), err: err}
		}
		return writeResp{n: len(p)}
	}

	return writeResp{err: result.Err}
}

func (f *File) ensureData() error {
	if !f.HasData {
		req := model.NewGetBlockReq(f.Block.Id)
		resp := f.FileSystem.fetchBlock(req)
		if resp.Err == nil {
			f.Block = resp.Block
			f.HasData = true
		} else {
			return resp.Err
		}
	}
	return nil
}

func (f *File) Name() string {
	name, err := f.Path.head()
	if err != nil {
		return ""
	}
	return string(name)
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
	return f.ModeValue.IsDir()
}

func (f *File) Sys() any {
	return nil
}
