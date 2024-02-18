package nodes

import (
	"github.com/google/uuid"
)

type Id string

func NewNodeId() Id {
	idValue := uuid.New()
	return Id(idValue.String())
}

type Slice []Id

func (p Slice) Len() int           { return len(p) }
func (p Slice) Less(i, j int) bool { return p[i] < p[j] }
func (p Slice) Swap(i, j int)      { p[i], p[j] = p[j], p[i] }
