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
	"bytes"
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

type StoredHash string
type Op int

const (
	Add Op = iota
	Remove
)

type StoredHashMessage struct {
	Op   Op
	Hash StoredHash
}

func (m StoredHashMessage) Serialize() []byte {
	op := model.IntToBytes(uint32(m.Op))
	hash := model.StringToBytes(string(m.Hash))
	return bytes.Join([][]byte{op, hash}, []byte{})
}

func toStoredHashMessage(data []byte) StoredHashMessage {
	op, remainder := model.IntFromBytes(data)
	hash, _ := model.StringFromBytes(remainder)
	return StoredHashMessage{
		Op:   Op(op),
		Hash: StoredHash(hash),
	}
}

func New(
	path Path,
	id model.NodeId,
	diskId model.DiskId,
	mgrDiskWrites chan model.WriteRequest,
	mgrDiskReads chan model.ReadRequest,
	mgrDiskBroadcast chan model.Broadcast,
	diskMgrWrites chan model.WriteResult,
	diskMgrReads chan model.ReadResult,
	diskMgrBroadcast chan model.Broadcast,
	ctx context.Context,
) Disk {
	p := Disk{
		path:         path,
		id:           id,
		diskId:       diskId,
		inWrites:     mgrDiskWrites,
		inReads:      mgrDiskReads,
		inBroadcast:  mgrDiskBroadcast,
		outReads:     diskMgrReads,
		outWrites:    diskMgrWrites,
		outBroadcast: diskMgrBroadcast,
		storedHashes: listStoredHashes(path),
		ctx:          ctx,
	}
	go p.consumeChannels()
	return p
}

func listStoredHashes(path Path) set.Set[StoredHash] {
	result := set.NewSet[StoredHash]()
	for _, fileName := range path.ListDirFiles() {
		result.Add(StoredHash(fileName))
	}
	return result
}

type Disk struct {
	path         Path
	id           model.NodeId
	diskId       model.DiskId
	outReads     chan model.ReadResult
	outWrites    chan model.WriteResult
	outBroadcast chan model.Broadcast
	inWrites     chan model.WriteRequest
	inReads      chan model.ReadRequest
	inBroadcast  chan model.Broadcast
	storedHashes set.Set[StoredHash]
	ctx          context.Context
}

func (d *Disk) Id() model.DiskId { return d.diskId }

func (d *Disk) consumeChannels() {
	for {
		select {
		case <-d.ctx.Done():
			return
		case s := <-d.inWrites:
			data := s.Data()
			err := d.path.Save(data)
			if err == nil {
				wr := model.NewWriteResultOk(s.Data().Ptr, s.Caller(), s.ReqId())
				hash := StoredHash(data.Ptr.FileName())
				msg := StoredHashMessage{Op: Add, Hash: hash}.Serialize()
				broadcast := model.NewBroadcast(msg, model.DiskDest)
				chanutil.Send(d.ctx, d.outBroadcast, broadcast, "disk: broadcast add")
				chanutil.Send(d.ctx, d.outWrites, wr, "disk: save success")
			} else {
				wr := model.NewWriteResultErr(err.Error(), s.Caller(), s.ReqId())
				chanutil.Send(d.ctx, d.outWrites, wr, "disk: save failure")
			}
		case r := <-d.inReads:
			if len(r.Ptrs()) == 0 {
				rr := model.NewReadResultErr("no pointers in read request", r.Caller(), r.GetBlockId(), r.BlockId())
				chanutil.Send(d.ctx, d.outReads, rr, "disk: no pointers in read request")
			} else {
				data, err := d.path.Read(r.Ptrs()[0])
				if err == nil {
					rr := model.NewReadResultOk(r.Caller(), r.Ptrs()[1:], data, r.GetBlockId(), r.BlockId())
					chanutil.Send(d.ctx, d.outReads, rr, "disk: read success")
				} else {
					rr := model.NewReadResultErr(err.Error(), r.Caller(), r.GetBlockId(), r.BlockId())
					chanutil.Send(d.ctx, d.outReads, rr, "disk: read failure")
				}
			}
		case b := <-d.inBroadcast:
			msg := toStoredHashMessage(b.Msg())
			switch msg.Op {
			case Add:
				d.storedHashes.Add(msg.Hash)
			case Remove:
				d.storedHashes.Remove(msg.Hash)
			default:
				log.Panicf("Unknown disk broadcast with op %d", msg.Op)
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

func (p *Path) ListDirFiles() []string {
	result := []string{}
	files, err := p.ops.ReadDir(p.raw)
	if err != nil {
		log.Panic("No dir here")
		return []string{}
	}
	for _, file := range files {
		if file.Type().IsRegular() {
			result = append(result, file.Name())
		}
	}
	return result
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
