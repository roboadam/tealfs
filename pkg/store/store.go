package store

import (
	"encoding/hex"
	"github.com/google/uuid"
	"os"
	"path/filepath"
	h "tealfs/pkg/hash"
	"tealfs/pkg/util"
	"time"
)

type Path struct {
	id  PathId
	raw string
}

type PathId struct {
	value string
}

type Paths struct {
	paths map[PathId]Path
	adds  chan Path
	keys  chan chan []PathId
	saves chan struct {
		pathId PathId
		hash   h.Hash
		data   []byte
	}
	reads chan struct {
		pathId PathId
		hash   h.Hash
		value  chan []byte
	}
}

func (ps *Paths) Add(p Path) {
	ps.adds <- p
}

func (ps *Paths) Keys() []PathId {
	response := make(chan []PathId)
	ps.keys <- response
	return <-response
}

func (ps *Paths) Save(id PathId, hash h.Hash, data []byte) {
	ps.saves <- struct {
		pathId PathId
		hash   h.Hash
		data   []byte
	}{pathId: id, hash: hash, data: data}
}

func (ps *Paths) Read(id PathId, hash h.Hash) []byte {
	value := make(chan []byte)
	ps.reads <- struct {
		pathId PathId
		hash   h.Hash
		value  chan []byte
	}{pathId: id, hash: hash, value: value}
	return <-value
}

func (ps *Paths) consumeChannels() {
	for {
		select {
		case p := <-ps.adds:
			ps.paths[p.id] = p
		case k := <-ps.keys:
			k <- util.Keys(ps.paths)
		case s := <-ps.saves:
			path := ps.paths[s.pathId]
			path.Save(s.hash, s.data)
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
		id:  PathId{value: uuid.New().String()},
	}
}

func NewPaths() Paths {
	p := Paths{
		paths: make(map[PathId]Path),
		adds:  make(chan Path),
		keys:  make(chan chan []PathId),
		saves: make(chan struct {
			pathId PathId
			hash   h.Hash
			data   []byte
		}),
		reads: make(chan struct {
			pathId PathId
			hash   h.Hash
			value  chan []byte
		}),
	}
	go p.consumeChannels()
	return p
}
