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
)

type Path struct {
	raw string
	ops FileOps
}

func New(
	path Path, id model.NodeId,
	mgrDiskWrites chan model.WriteRequest,
	mgrDiskReads chan model.ReadRequest,
	diskMgrWrites chan model.WriteResult,
	diskMgrReads chan model.ReadResult,
	ctx context.Context,
) Disk {
	p := Disk{
		path:      path,
		id:        id,
		inWrites:  mgrDiskWrites,
		inReads:   mgrDiskReads,
		outReads:  diskMgrReads,
		outWrites: diskMgrWrites,
	}
	go p.consumeChannels(ctx)
	return p
}

type Disk struct {
	path      Path
	id        model.NodeId
	outReads  chan model.ReadResult
	outWrites chan model.WriteResult
	inWrites  chan model.WriteRequest
	inReads   chan model.ReadRequest
}

func (d *Disk) consumeChannels(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			return
		case s := <-d.inWrites:
			err := d.path.Save(s.Data())
			if err == nil {
				wr := model.NewWriteResultOk(s.Data().Ptr, s.Caller(), s.ReqId())
				chanutil.Send(d.outWrites, wr, "disk: save success")
			} else {
				wr := model.NewWriteResultErr(err.Error(), s.Caller(), s.ReqId())
				chanutil.Send(d.outWrites, wr, "disk: save failure")
			}
		case r := <-d.inReads:
			if len(r.Ptrs()) == 0 {
				rr := model.NewReadResultErr("no pointers in read request", r.Caller(), r.GetBlockId(), r.BlockId())
				chanutil.Send(d.outReads, rr, "disk: no pointers in read request")
			} else {
				data, err := d.path.Read(r.Ptrs()[0])
				if err == nil {
					rr := model.NewReadResultOk(r.Caller(), r.Ptrs()[1:], data, r.GetBlockId(), r.BlockId())
					chanutil.Send(d.outReads, rr, "disk: read success")
				} else {
					rr := model.NewReadResultErr(err.Error(), r.Caller(), r.GetBlockId(), r.BlockId())
					chanutil.Send(d.outReads, rr, "disk: read failure")
				}
			}
		}
	}
}

func (p *Path) Save(rawData model.RawData) error {
	filePath := filepath.Join(p.raw, rawData.Ptr.FileName())
	return p.ops.WriteFile(filePath, rawData.Data)
}

func (p *Path) Read(ptr model.DiskPointer) (model.RawData, error) {
	filePath := filepath.Join(p.raw, ptr.FileName())
	result, err := p.ops.ReadFile(filePath)
	if err != nil && errors.Is(err, fs.ErrNotExist) {
		return model.RawData{Ptr: ptr, Data: []byte{}}, nil
	}
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
