package store

import (
	"github.com/google/uuid"
	"path/filepath"
	"tealfs/pkg/util"
)

type Path struct {
	id  uuid.UUID
	raw string
}

type Paths struct {
	paths util.Set[Path]
}

func NewPath(rawPath string) Path {
	return Path{
		id:  uuid.New(),
		raw: filepath.Clean(rawPath),
	}
}

func NewPaths() Paths {
	return Paths{paths: util.NewSet[Path]()}
}
