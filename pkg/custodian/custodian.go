// Copyright (C) 2025 Adam Hess
//
// This program is free software: you can redistribute it and/or modify it under
// the terms of the GNU Affero General Public License as published by the Free
// Software Foundation, version 3.
//
// This program is distributed in the hope that it will be useful, but WITHOUT
// ANY WARRANTY; without even the implied warranty of MERCHANTABILITY or
// FITNESS FOR A PARTICULAR PURPOSE. See the GNU Affero General Public License
// for more details.
//
// You should have received a copy of the GNU Affero General Public License
// along with this program. If not, see <https://www.gnu.org/licenses/>.

package custodian

import (
	"context"
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
	Commands       chan Command
	globalBlockIds set.Set[model.BlockId]
}

func NewCustodian(chanSize int) *Custodian {
	c := Custodian{globalBlockIds: set.NewSet[model.BlockId]()}
	return &c
}

func (c *Custodian) Start(ctx context.Context) {
	go c.processCommands(ctx)
}

func (c *Custodian) processCommands(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			return
		case cmd := <-c.Commands:
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
}
