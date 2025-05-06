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

const BytesPerBlock = 1_048_576

type File struct {
	SizeValue  int64
	ModeValue  fs.FileMode
	Modtime    time.Time
	Position   int64
	Block      []model.Block
	HasData    []bool
	Path       Path
	FileSystem *FileSystem
}

func (f *File) ToBytes() []byte {
	value := model.IntToBytes(uint32(f.SizeValue))
	value = append(value, model.IntToBytes(uint32(f.ModeValue))...)
	value = append(value, model.IntToBytes(uint32(f.Modtime.Unix()))...)
	value = append(value, model.IntToBytes(uint32(len(f.Block)))...)
	for _, b := range f.Block {
		value = append(value, model.StringToBytes(string(b.Id))...)
	}
	value = append(value, model.StringToBytes(string(f.Path.toName()))...)
	return value
}

func FileFromBytes(raw []byte, fileSystem *FileSystem) (File, []byte, error) {
	size, remainder := model.IntFromBytes(raw)
	mode, remainder := model.IntFromBytes(remainder)
	modtimeRaw, remainder := model.IntFromBytes(remainder)
	blockLen, remainder := model.IntFromBytes(remainder)
	blocks := make([]model.Block, blockLen)
	for i := range blockLen {
		var blockId string
		blockId, remainder = model.StringFromBytes(remainder)
		blocks[i].Id = model.BlockId(blockId)
		blocks[i].Data = []byte{}
		blocks[i].Type = model.Mirrored
	}
	rawPath, remainder := model.StringFromBytes(remainder)

	path, err := PathFromName(rawPath)
	if err != nil {
		return File{}, nil, err
	}
	modTime := time.Unix(int64(modtimeRaw), 0)
	return File{
		SizeValue:  int64(size),
		ModeValue:  fs.FileMode(mode),
		Modtime:    modTime,
		Position:   0,
		Block:      blocks,
		HasData:    make([]bool, blockLen),
		Path:       path,
		FileSystem: fileSystem,
	}, remainder, nil
}

type closeReq struct {
	f    *File
	resp chan closeResp
}
type closeResp struct{ err error }

func (f *File) Close() error {
	req := closeReq{f: f, resp: make(chan closeResp)}
	chanutil.Send(f.FileSystem.Ctx, f.FileSystem.closeReq, req, "close")
	resp := <-req.resp
	return resp.err
}
func closeF(req closeReq) closeResp {
	f := req.f
	f.Position = 0
	for i := range len(f.Block) {
		f.Block[i].Data = []byte{}
		f.HasData[i] = false
	}
	return closeResp{}
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
	chanutil.Send(f.FileSystem.Ctx, f.FileSystem.readReq, req, "read")
	resp := <-req.resp
	return resp.n, resp.err
}

func (f *File) eof() bool {
	return f.isPositionEof(f.Position)
}

func (f *File) isPositionEof(position int64) bool {
	return position >= f.SizeValue
}

func positionToBlockIndexAndOffset(position int64) (int, int64) {
	index := int(position / BytesPerBlock)
	offset := position - int64(index*BytesPerBlock)
	return index, offset
}

func read(req readReq) readResp {
	f := req.f
	p := req.p
	err := f.ensureData()
	if err != nil {
		log.Warn("Error reading data for ", f.Path.toName())
		return readResp{err: err}
	}

	if f.eof() {
		log.Warn("EOF reading data for ", f.Path.toName())
		return readResp{err: io.EOF}
	}

	start := f.Position
	end := f.Position + int64(len(p))
	if f.isPositionEof(end) {
		end = f.SizeValue
	}

	// Phase 1: identify the blocks we need locally
	firstIndex, firstOffset := positionToBlockIndexAndOffset(start)
	lastIndex, lastOffset := positionToBlockIndexAndOffset(end)

	// Phase 2: load those blocks
	for i := firstIndex; i <= lastIndex; i++ {
		err := f.ensureDataForIndex(i)
		if err != nil {
			return readResp{err: err}
		}
	}

	if firstIndex == lastIndex {
		// Phase 3a: if within one block
		//           return data[block][offsetStart:offsetEnd]
		copy(p, f.Block[firstIndex].Data[firstOffset:lastOffset])
	} else {
		// Phase 3b: if in non-adjacent blocks
		//           return data[blockStart][offsetStart:] + ... + data[blocksInMiddle] + ... + data[blockEnd][:offsetEnd]
		bytesRead := 0
		src := f.Block[firstIndex].Data[firstOffset:]
		copy(p, src)
		bytesRead += len(src)
		for i := firstIndex + 1; i < lastIndex; i++ {
			copy(p[bytesRead:], f.Block[i].Data)
			bytesRead += BytesPerBlock
		}
		copy(p[bytesRead:], f.Block[lastIndex].Data[:lastOffset])
	}

	bytesRead := int(end - start)
	f.Position += int64(bytesRead)
	return readResp{n: bytesRead}
}

type seekReq struct {
	offset int64
	whence int
	f      *File
	resp   chan seekResp
}
type seekResp struct {
	pos int64
	err error
}

func (f *File) Seek(offset int64, whence int) (int64, error) {
	req := seekReq{
		offset: offset,
		whence: whence,
		f:      f,
		resp:   make(chan seekResp),
	}
	chanutil.Send(f.FileSystem.Ctx, f.FileSystem.seekReq, req, "seek")
	resp := <-req.resp
	return resp.pos, resp.err
}

func seek(req seekReq) seekResp {
	whence := req.whence
	offset := req.offset
	f := req.f
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
		return seekResp{pos: f.Position, err: errors.New("negative seek")}
	}
	f.Position = int64(newPosition)
	return seekResp{pos: f.Position}
}

type readdirReq struct {
	f     *File
	count int
	resp  chan readdirResp
}
type readdirResp struct {
	infos []fs.FileInfo
	err   error
}

func (f *File) Readdir(count int) ([]fs.FileInfo, error) {
	req := readdirReq{
		f:     f,
		count: count,
		resp:  make(chan readdirResp),
	}
	chanutil.Send(f.FileSystem.Ctx, f.FileSystem.readdirReq, req, "readdir")
	resp := <-req.resp
	return resp.infos, resp.err
}

func readdir(req readdirReq) readdirResp {
	f := req.f
	count := req.count
	if count < 0 {
		return readdirResp{err: errors.New("negative dir count requested")}
	}
	children := f.FileSystem.immediateChildren(f.Path)[count:]
	result := make([]fs.FileInfo, 0, len(children))
	for _, child := range children {
		result = append(result, child)
	}
	return readdirResp{infos: result}
}

type statReq struct {
	f    *File
	resp chan statResp
}
type statResp struct {
	info fs.FileInfo
	err  error
}

func (f *File) Stat() (fs.FileInfo, error) {
	req := statReq{
		f:    f,
		resp: make(chan statResp),
	}
	chanutil.Send(f.FileSystem.Ctx, f.FileSystem.statReq, req, "stat")
	resp := <-req.resp
	return resp.info, resp.err
}
func stat(req statReq) statResp {
	return statResp{info: req.f}
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
	chanutil.Send(f.FileSystem.Ctx, f.FileSystem.writeReq, req, "write")
	resp := <-req.resp
	return resp.n, resp.err
}

func (f *File) growFile(byteCount int64) {
	if byteCount > 0 {

	}
}

func write(wreq writeReq) writeResp {
	f := wreq.f
	p := wreq.p

	start := f.Position
	end := f.Position + int64(len(p))

	// Phase 1: identify the blocks we need locally
	firstIndex, firstOffset := positionToBlockIndexAndOffset(start)
	lastIndex, _ := positionToBlockIndexAndOffset(end)

	// Phase 2: create blocks
	if lastIndex >= len(f.Block) {
		numBlocksToCreate := lastIndex - len(f.Block) + 1
		for range numBlocksToCreate {
			f.Block = append(f.Block, model.Block{
				Id:   model.NewBlockId(),
				Type: model.Mirrored,
				Data: []byte{},
			})
		}
	}

	// Phase 3: load those blocks
	for i := firstIndex; i <= lastIndex; i++ {
		err := f.ensureDataForIndex(i)
		if err != nil {
			return writeResp{err: err}
		}
	}

	bytesWritten := 0
	// Phase 4: if in non-adjacent blocks
	//           return data[blockStart][offsetStart:] + ... + data[blocksInMiddle] + ... + data[blockEnd][:offsetEnd]
	bytesWritten += copy(f.Block[firstIndex].Data[firstOffset:], p)
	req := model.NewPutBlockReq(f.Block[firstIndex])
	result := f.FileSystem.pushBlock(req)
	if result.Err != nil {
		return writeResp{err: result.Err}
	}
	for i := firstIndex + 1; i <= lastIndex; i++ {
		bytesWritten += copy(f.Block[i].Data, p[bytesWritten:])
		req := model.NewPutBlockReq(f.Block[i])
		result := f.FileSystem.pushBlock(req)
		if result.Err != nil {
			return writeResp{err: result.Err}
		}
	}

	f.Position += int64(bytesWritten)

	err := f.FileSystem.persistFileIndexAndBroadcast(f, upsertFile)
	if err != nil {
		return writeResp{n: bytesWritten, err: err}
	}
	return writeResp{n: bytesWritten}
}

func (f *File) ensureData() error {
	currentBlockIndex := int(f.Position / BytesPerBlock)
	return f.ensureDataForIndex(currentBlockIndex)
}

func (f *File) ensureDataForIndex(index int) error {
	if !f.HasData[index] {
		req := model.NewGetBlockReq(f.Block[index].Id)
		resp := f.FileSystem.fetchBlock(req)
		if resp.Err == nil {
			f.Block[index] = resp.Block
			f.HasData[index] = true
		} else {
			return resp.Err
		}
	}
	return nil
}

type nameReq struct {
	f    *File
	resp chan nameResp
}
type nameResp struct {
	name string
}

func (f *File) Name() string {
	req := nameReq{
		f:    f,
		resp: make(chan nameResp),
	}
	chanutil.Send(f.FileSystem.Ctx, f.FileSystem.nameReq, req, "name")
	resp := <-req.resp
	return resp.name
}
func name(req nameReq) nameResp {
	f := req.f
	name, err := f.Path.head()
	if err != nil {
		return nameResp{
			name: "",
		}
	}
	return nameResp{name: string(name)}
}

type sizeReq struct {
	f    *File
	resp chan sizeResp
}
type sizeResp struct {
	size int64
}

func (f *File) Size() int64 {
	req := sizeReq{
		f:    f,
		resp: make(chan sizeResp),
	}
	chanutil.Send(f.FileSystem.Ctx, f.FileSystem.sizeReq, req, "size")
	resp := <-req.resp
	return resp.size
}
func size(req sizeReq) sizeResp {
	f := req.f
	return sizeResp{size: f.SizeValue}
}

type modeReq struct {
	f    *File
	resp chan modeResp
}
type modeResp struct {
	mode fs.FileMode
}

func (f *File) Mode() fs.FileMode {
	req := modeReq{
		f:    f,
		resp: make(chan modeResp),
	}
	chanutil.Send(f.FileSystem.Ctx, f.FileSystem.modeReq, req, "mode")
	resp := <-req.resp
	return resp.mode
}
func mode(req modeReq) modeResp {
	f := req.f
	return modeResp{mode: f.ModeValue}
}

type modtimeReq struct {
	f    *File
	resp chan modtimeResp
}
type modtimeResp struct {
	time time.Time
}

func (f *File) ModTime() time.Time {
	req := modtimeReq{
		f:    f,
		resp: make(chan modtimeResp),
	}
	chanutil.Send(f.FileSystem.Ctx, f.FileSystem.modtimeReq, req, "modtime")
	resp := <-req.resp
	return resp.time
}
func modtime(req modtimeReq) modtimeResp {
	f := req.f
	return modtimeResp{time: f.Modtime}
}

type isdirReq struct {
	f    *File
	resp chan isdirResp
}
type isdirResp struct {
	is bool
}

func (f *File) IsDir() bool {
	req := isdirReq{
		f:    f,
		resp: make(chan isdirResp),
	}
	chanutil.Send(f.FileSystem.Ctx, f.FileSystem.isdirReq, req, "isdir")
	resp := <-req.resp
	return resp.is
}
func isdir(req isdirReq) isdirResp {
	f := req.f
	return isdirResp{is: f.ModeValue.IsDir()}
}

type sysReq struct {
	f    *File
	resp chan sysResp
}
type sysResp struct {
	whatever any
}

func (f *File) Sys() any {
	req := sysReq{
		f:    f,
		resp: make(chan sysResp),
	}
	chanutil.Send(f.FileSystem.Ctx, f.FileSystem.sysReq, req, "sys")
	resp := <-req.resp
	return resp.whatever
}
func sys(_ sysReq) sysResp {
	return sysResp{}
}
