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

package datalayer

import (
	"tealfs/pkg/model"
)

type StateHandler struct {
	outSaveRequest   chan<- saveRequest
	outDeleteRequest chan<- deleteRequest

	state      State

	MainNodeId model.NodeId
	MyNodeId   model.NodeId
	NodeConnMap *model.NodeConnectionMapper
}

func (s *StateHandler) SetDiskSpace(d dest, space int) {

}

func (s *StateHandler) Saved(blockId model.BlockId, d dest) {

}

func (s *StateHandler) Deleted(b model.BlockId, d dest) {

}
