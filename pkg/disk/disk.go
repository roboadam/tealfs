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

package disk

import (
	"context"
	"errors"
	"io/fs"
	"path/filepath"
	"tealfs/pkg/chanutil"
	"tealfs/pkg/model"
	"tealfs/pkg/set"

	log "github.com/sirupsen/logrus"
)

type Path struct {
	raw string
	ops FileOps
}

func New(
	path Path,
	id model.NodeId,
	diskId model.DiskId,
	ctx context.Context,
) Disk {
	p := Disk{
		path:      path,
		id:        id,
		diskId:    diskId,
		InWrites:  make(chan model.WriteRequest, 1),
		InReads:   make(chan model.ReadRequest, 1),
		InExists:  make(chan ExistsReq, 1),
		OutReads:  make(chan model.ReadResult, 1),
		OutWrites: make(chan model.WriteResult, 1),
		inGet: make(chan struct {
			blockId model.BlockId
			resp    chan struct {
				data []byte
				ok   bool
			}
		}, 1),
		inSave: make(chan struct {
			data    []byte
			blockId model.BlockId
			resp    chan bool
		}),
		ctx:        ctx,
	}
	go p.consumeChannels()
	return p
}

type Disk struct {
	path       Path
	id         model.NodeId
	diskId     model.DiskId
	OutReads   chan model.ReadResult
	OutWrites  chan model.WriteResult
	InWrites   chan model.WriteRequest
	InReads    chan model.ReadRequest
	InListIds  chan struct{}
	OutListIds chan set.Set[model.BlockId]
	InDelete   chan model.BlockId
	InExists   chan ExistsReq
	inGet      chan struct {
		blockId model.BlockId
		resp    chan struct {
			data []byte
			ok   bool
		}
	}
	inSave chan struct {
		data    []byte
		blockId model.BlockId
		resp    chan bool
	}
	ctx context.Context
}

type ExistsReq struct {
	BlockId model.BlockId
	Resp    chan bool
}

func (d *Disk) Id() model.DiskId { return d.diskId }

func (d *Disk) Get(blockId model.BlockId) ([]byte, bool) {
	resp := make(chan struct {
		data []byte
		ok   bool
	})
	d.inGet <- struct {
		blockId model.BlockId
		resp    chan struct {
			data []byte
			ok   bool
		}
	}{
		blockId: blockId,
		resp:    resp,
	}
	result := <-resp
	return result.data, result.ok
}

func (d *Disk) Save(data []byte, blockId model.BlockId) bool {
	resp := make(chan bool)
	d.inSave <- struct {
		data    []byte
		blockId model.BlockId
		resp    chan bool
	}{
		data:    data,
		blockId: blockId,
		resp:    resp,
	}
	return <-resp
}

func (d *Disk) consumeChannels() {
	for {
		select {
		case <-d.ctx.Done():
			return
		case get := <-d.inGet:
			data, err := d.path.ReadDirect(model.DiskPointer{
				NodeId:   d.id,
				Disk:     d.diskId,
				FileName: string(get.blockId),
			})
			ok := err == nil
			get.resp <- struct {
				data []byte
				ok   bool
			}{
				data: data.Data,
				ok:   ok,
			}
		case save := <-d.inSave:
			err := d.path.Save(model.RawData{
				Ptr: model.DiskPointer{
					NodeId:   d.id,
					Disk:     d.diskId,
					FileName: string(save.blockId),
				},
				Data: save.data,
			})
			save.resp <- err == nil
		case s := <-d.InWrites:
			err := d.path.Save(s.Data)
			if err == nil {
				wr := model.NewWriteResultOk(s.Data.Ptr, s.Caller, s.ReqId)
				chanutil.Send(d.ctx, d.OutWrites, wr, "disk: save success")
			} else {
				wr := model.NewWriteResultErr(err.Error(), s.Caller, s.ReqId)
				chanutil.Send(d.ctx, d.OutWrites, wr, "disk: save failure")
			}
		case r := <-d.InReads:
			if len(r.Ptrs) == 0 {
				rr := model.NewReadResultErr("no pointers in read request", r.Caller, r.ReqId, r.BlockId)
				chanutil.Send(d.ctx, d.OutReads, rr, "disk: no pointers in read request")
			} else {
				data, err := d.path.ReadOrEmpty(r.Ptrs[0])
				if err == nil {
					rr := model.NewReadResultOk(r.Caller, r.Ptrs[1:], data, r.ReqId, r.BlockId)
					chanutil.Send(d.ctx, d.OutReads, rr, "disk: read success")
				} else {
					rr := model.NewReadResultErr(err.Error(), r.Caller, r.ReqId, r.BlockId)
					chanutil.Send(d.ctx, d.OutReads, rr, "disk: read failure")
				}
			}
		case <-d.InListIds:
			allIds := set.NewSet[model.BlockId]()
			files, err := d.path.ops.ListFiles(d.path.raw)
			if err == nil {
				for _, f := range files {
					allIds.Add(model.BlockId(f))
				}
			}
			d.OutListIds <- allIds
		case idToDelete := <-d.InDelete:
			filePath := filepath.Join(d.path.raw, string(idToDelete))
			err := d.path.ops.Remove(filePath)
			if err != nil {
				log.Warn("Error deleting file")
			}
		case req := <-d.InExists:
			filePath := filepath.Join(d.path.raw, string(req.BlockId))
			req.Resp <- d.path.ops.Exists(filePath)
		}
	}
}

func (p *Path) Save(rawData model.RawData) error {
	filePath := filepath.Join(p.raw, rawData.Ptr.FileName)
	return p.ops.WriteFile(filePath, rawData.Data)
}

func (p *Path) ReadOrEmpty(ptr model.DiskPointer) (model.RawData, error) {
	data, err := p.ReadDirect(ptr)
	if err != nil && errors.Is(err, fs.ErrNotExist) {
		return model.RawData{Ptr: ptr, Data: []byte{}}, nil
	}
	return data, err
}

func (p *Path) ReadDirect(ptr model.DiskPointer) (model.RawData, error) {
	filePath := filepath.Join(p.raw, ptr.FileName)
	result, err := p.ops.ReadFile(filePath)
	return model.RawData{Ptr: ptr, Data: result}, err
}

func NewPath(rawPath string, ops FileOps) Path {
	return Path{
		raw: filepath.Clean(rawPath),
		ops: ops,
	}
}

func (p *Path) String() string {
	return p.raw
}
