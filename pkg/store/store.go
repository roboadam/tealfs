package store

import (
	"encoding/hex"
	"github.com/google/uuid"
	"os"
	"path/filepath"
	h "tealfs/pkg/hash"
	"time"
)

type Path struct {
	raw string
}

type Id struct {
	value string
}

type Store struct {
	path  Path
	id    Id
	saves chan struct {
		hash h.Hash
		data []byte
	}
	reads chan struct {
		hash  h.Hash
		value chan []byte
	}
}

func (ps *Store) Save(hash h.Hash, data []byte) {
	ps.saves <- struct {
		hash h.Hash
		data []byte
	}{hash: hash, data: data}
}

func (ps *Store) Read(hash h.Hash) []byte {
	value := make(chan []byte)
	ps.reads <- struct {
		hash  h.Hash
		value chan []byte
	}{hash: hash, value: value}
	return <-value
}

func (ps *Store) consumeChannels() {
	for {
		select {
		case s := <-ps.saves:
			ps.path.Save(s.hash, s.data)
		case r := <-ps.reads:
			data := ps.path.Read(r.hash)
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

func New(path Path) Store {
	p := Store{
		path: path,
		id:   Id{value: uuid.New().String()},
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

func (p Id) String() string {
	return p.value
}

type Slice []Id

func (p Slice) Len() int           { return len(p) }
func (p Slice) Less(i, j int) bool { return p[i].String() < p[j].String() }
func (p Slice) Swap(i, j int)      { p[i], p[j] = p[j], p[i] }
