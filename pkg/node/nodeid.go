package node

import (
	"fmt"
	"os"

	"github.com/google/uuid"
)

type NodeId struct {
	value uuid.UUID
}

func (nodeId NodeId) String() string {
	return nodeId.value.String()
}

func NewNodeId() NodeId {
	uuid, err := uuid.NewUUID()
	if err != nil {
		fmt.Println("Error generating UUID:", err)
		os.Exit(1)
	}

	return NodeId{
		value: uuid,
	}
}
