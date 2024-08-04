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
	"time"
)

type Path struct {
	raw string
}

type Disk struct {
	path  Path
	id    model.NodeId
	saves chan struct {
		hash h.Hash
		data []byte
	}
	reads chan struct {
		hash  h.Hash
		value chan []byte
	}
	OutReads  chan model.ReadResult
	OutWrites chan model.WriteResult
	InWrites  chan model.Block
	InReads   chan model.ReadRequest
}

func (d *Disk) Save(hash h.Hash, data []byte) {
	d.saves <- struct {
		hash h.Hash
		data []byte
	}{hash: hash, data: data}
}

func (d *Disk) Read(hash h.Hash) []byte {
	value := make(chan []byte)
	d.reads <- struct {
		hash  h.Hash
		value chan []byte
	}{hash: hash, value: value}
	return <-value
}

func (d *Disk) consumeChannels() {
	for {
		select {
		case s := <-d.saves:
			d.path.Save(s.hash, s.data)
		case r := <-d.reads:
			data := d.path.Read(r.hash)
			r.value <- data
		}
	}
}

func (p *Path) Save(hash h.Hash, data []byte) {
	for {
		hashString := hex.EncodeToString(hash.Value)
		filePath := filepath.Join(p.raw, hashString)
		err := os.WriteFile(filePath, data, 0644)
		if err == nil {
			return
		}
		time.Sleep(time.Second)
	}
}

func (p *Path) Read(hash h.Hash) []byte {
	count := 0
	for {
		count++
		hashString := hex.EncodeToString(hash.Value)
		filePath := filepath.Join(p.raw, hashString)
		data, err := os.ReadFile(filePath)
		if err == nil {
			return data
		}
		if count > 5 {
			return make([]byte, 0)
		}
		time.Sleep(time.Second)
	}
}

func NewPath(rawPath string) Path {
	return Path{
		raw: filepath.Clean(rawPath),
	}
}

func New(path Path, id model.NodeId) Disk {
	p := Disk{
		id:   id,
		path: path,
		saves: make(chan struct {
			hash h.Hash
			data []byte
		}),
		reads: make(chan struct {
			hash  h.Hash
			value chan []byte
		}),
	}
	go p.consumeChannels()
	return p
}

func (p *Path) String() string {
	return p.raw
}
