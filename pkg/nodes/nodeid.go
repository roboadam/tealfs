package nodes

import (
	"github.com/google/uuid"
)

type Id struct {
	value string
}

func (nodeId Id) String() string {
	return nodeId.value
}

func IdFromRaw(rawId string) Id {
	return Id{
		value: rawId,
	}
}

func NewNodeId() Id {
	idValue := uuid.New()

	return Id{
		value: idValue.String(),
	}
}

type Slice []Id

func (p Slice) Len() int           { return len(p) }
func (p Slice) Less(i, j int) bool { return p[i].String() < p[j].String() }
func (p Slice) Swap(i, j int)      { p[i], p[j] = p[j], p[i] }
