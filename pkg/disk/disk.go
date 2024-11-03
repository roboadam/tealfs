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

package disk

import (
	"errors"
	"fmt"
	"io/fs"
	"path/filepath"
	"tealfs/pkg/model"
)

type Path struct {
	raw string
	ops FileOps
}

func New(path Path, id model.NodeId,
	mgrDiskWrites chan model.WriteRequest,
	mgrDiskReads chan model.ReadRequest,
	diskMgrWrites chan model.WriteResult,
	diskMgrReads chan model.ReadResult) Disk {
	p := Disk{
		path:      path,
		id:        id,
		inWrites:  mgrDiskWrites,
		inReads:   mgrDiskReads,
		outReads:  diskMgrReads,
		outWrites: diskMgrWrites,
	}
	go p.consumeChannels()
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

func (d *Disk) consumeChannels() {
	for {
		select {
		case s := <-d.inWrites:
			fmt.Println("Disk In Write")
			err := d.path.Save(s.Block.Id, s.Block.Data)
			if err == nil {
				fmt.Println("Disk Out Write Success")
				d.outWrites <- model.WriteResult{
					Ok:      true,
					Caller:  s.Caller,
					BlockId: s.Block.Id,
				}
			} else {
				fmt.Println("Disk Out Write Failure")
				d.outWrites <- model.WriteResult{
					Ok:      false,
					Message: err.Error(),
					Caller:  s.Caller,
					BlockId: s.Block.Id,
				}
			}
		case r := <-d.inReads:
			fmt.Println("Disk In Read")
			data, err := d.path.Read(r.BlockId)
			if err == nil {
				fmt.Println("Disk Out Read Success")
				d.outReads <- model.ReadResult{
					Ok:     true,
					Caller: r.Caller,
					Block: model.Block{
						Id:   r.BlockId,
						Data: data,
					},
				}
			} else {
				fmt.Println("Disk Out Read Failure")
				d.outReads <- model.ReadResult{
					Ok:      false,
					Message: err.Error(),
					Caller:  r.Caller,
				}
			}
		}
	}
}

func (p *Path) Save(id model.BlockId, data []byte) error {
	filePath := filepath.Join(p.raw, string(id))
	return p.ops.WriteFile(filePath, data)
}

func (p *Path) Read(id model.BlockId) ([]byte, error) {
	filePath := filepath.Join(p.raw, string(id))
	result, err := p.ops.ReadFile(filePath)
	if err != nil && errors.Is(err, fs.ErrNotExist) {
		return []byte{}, nil
	}
	return result, err
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
