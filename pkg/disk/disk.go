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
	"encoding/hex"
	"os"
	"path/filepath"
	h "tealfs/pkg/hash"
	"tealfs/pkg/model"
)

type Path struct {
	raw string
}

func New(path Path, id model.NodeId,
	mgrDiskWrites chan model.Block,
	mgrDiskReads chan model.ReadRequest,
	diskMgrWrites chan model.WriteResult,
	diskMgrReads chan model.ReadResult) Disk {
	p := Disk{
		path:      path,
		id:        id,
		outReads:  make(chan model.ReadResult),
		outWrites: make(chan model.WriteResult),
		inWrites:  mgrDiskWrites,
		inReads:   make(chan model.ReadRequest),
	}
	go p.consumeChannels()
	return p
}

type Disk struct {
	path      Path
	id        model.NodeId
	outReads  chan model.ReadResult
	outWrites chan model.WriteResult
	inWrites  chan model.Block
	inReads   chan model.ReadRequest
}

func (d *Disk) consumeChannels() {
	for {
		select {
		case s := <-d.inWrites:
			err := d.path.Save(s.Hash, s.Data)
			if err == nil {
				d.outWrites <- model.WriteResult{Ok: true}
			} else {
				d.outWrites <- model.WriteResult{Ok: false, Message: err.Error()}
			}
		case r := <-d.inReads:
			data, err := d.path.Read(r.BlockId)
			if err == nil {
				d.outReads <- model.ReadResult{
					Ok:     true,
					Caller: r.Caller,
					Block: model.Block{
						Id:   r.BlockId,
						Data: data,
						Hash: h.ForData(data),
					},
				}
			} else {
				d.outReads <- model.ReadResult{
					Ok:      false,
					Message: err.Error(),
					Caller:  r.Caller,
				}
			}
		}
	}
}

func (p *Path) Save(hash h.Hash, data []byte) error {
	hashString := hex.EncodeToString(hash.Value)
	filePath := filepath.Join(p.raw, hashString)
	return os.WriteFile(filePath, data, 0644)
}

func (p *Path) Read(id model.BlockId) ([]byte, error) {
	filePath := filepath.Join(p.raw, string(id))
	return os.ReadFile(filePath)
}

func NewPath(rawPath string) Path {
	return Path{
		raw: filepath.Clean(rawPath),
	}
}

func (p *Path) String() string {
	return p.raw
}
