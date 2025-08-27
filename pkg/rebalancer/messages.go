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

package rebalancer

import (
	"tealfs/pkg/model"
	"tealfs/pkg/set"
)

type AllBlockId string

type AllBlockIdReq struct {
	Caller model.NodeId
	Id     AllBlockId
}

func (a *AllBlockIdReq) Type() model.PayloadType {
	return model.AllBlockIdReqType
}

type AllBlockIdResp struct {
	Caller   model.NodeId
	BlockIds set.Set[model.BlockId]
	Id       AllBlockId
}

func (a *AllBlockIdResp) Type() model.PayloadType {
	return model.AllBlockIdRespType
}

type DeleteBlockId struct {
	BlockId model.BlockId
}

func (d *DeleteBlockId) Type() model.PayloadType {
	return model.DeleteBlockId
}
