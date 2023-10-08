package node

import (
	"fmt"
	"os"

	"github.com/google/uuid"
)

type Id struct {
	value uuid.UUID
}

func (nodeId Id) String() string {
	return nodeId.value.String()
}

func NewNodeId() Id {
	idValue, err := uuid.NewUUID()
	if err != nil {
		fmt.Println("Error generating UUID:", err)
		os.Exit(1)
	}

	return Id{
		value: idValue,
	}
}
