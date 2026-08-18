// Copyright (C) 2026 Adam Hess
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

package model

import (
	"encoding/gob"
)

type PayloadType uint16

func init() {
	gob.Register(&WriteRequest{})
	gob.Register(&IAm{})

}

type Payload2 interface {
	Destination() NodeId
}

type FetchBlockId string

type FetchBlockReq struct {
	Caller  NodeId
	BlockId BlockId
	Id      FetchBlockId
}

func (f *FetchBlockReq) Destination() NodeId {
	return ""
}

type FetchBlockCmd struct {
	Sources []NodeDisk
	BlockId BlockId
	Caller  NodeId
	Id      FetchBlockId
}

func (f *FetchBlockCmd) Destination() NodeId {
	return f.Sources[0].NodeId
}

type FetchBlockResp struct {
	Caller  NodeId
	Block   Block
	Id      FetchBlockId
	Success bool
	Msg     string
}

func (f *FetchBlockResp) Destination() NodeId {
	return f.Caller
}
