package store

import (
	"encoding/hex"
	"github.com/google/uuid"
	"os"
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
		id:  uuid.New(),
		raw: filepath.Clean(rawPath),
	}
}

func NewPaths() Paths {
	return Paths{paths: util.NewSet[Path]()}
}
