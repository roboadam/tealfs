package node

import (
	"fmt"
	"os"

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
	idValue, err := uuid.NewUUID()
	if err != nil {
		fmt.Println("Error generating UUID:", err)
		os.Exit(1)
	}

	return Id{
		value: idValue.String(),
	}
}
