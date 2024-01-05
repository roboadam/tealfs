package store

import (
	"encoding/hex"
	"github.com/google/uuid"
	"os"
	"path/filepath"
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
	keys  chan struct {
		response chan []PathId
	}
}

func (ps *Paths) Add(p Path) {
	ps.adds <- p
}

func (ps *Paths) Keys() []PathId {
	r := struct{ response chan []PathId }{response: make(chan []PathId)}
	ps.keys <- r
	return <-r.response
}

func (ps *Paths) consumeChannels() {
	for {
		select {
		case p := <-ps.adds:
			ps.paths[p.id] = p
		}
	}
}

func (p *Path) Save(hash []byte, data []byte) error {
	hashString := hex.EncodeToString(hash)
	filePath := filepath.Join(p.raw, hashString)
	err := os.WriteFile(filePath, data, 0644)
	if err != nil {
		return err
	}

	return nil
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
		keys:  make(chan struct{ response chan []PathId }),
	}
	go p.consumeChannels()
	return p
}
