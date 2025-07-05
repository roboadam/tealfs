package custodian

import (
	"tealfs/pkg/model"
	"tealfs/pkg/set"
)

type CommandType int

const (
	CommandTypeUnknown CommandType = iota
	CommandTypeAddBlock
	CommandTypeRemoveBlock
)

type Command struct {
	Type    CommandType
	BlockId model.BlockId
}

type Custodian struct {
	commands       chan Command
	globalBlockIds set.Set[model.BlockId]
}

func NewCustodian() *Custodian {
	c := Custodian{
		globalBlockIds: set.NewSet[model.BlockId](),
	}
	go c.processCommands()
	return &c
}

func (c *Custodian) processCommands() {
	for cmd := range c.commands {
		switch cmd.Type {
		case CommandTypeAddBlock:
			c.globalBlockIds.Add(cmd.BlockId)
		case CommandTypeRemoveBlock:
			c.globalBlockIds.Remove(cmd.BlockId)
		default:
			panic("unknown command type")
		}
	}
}
