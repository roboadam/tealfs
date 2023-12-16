package node

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
