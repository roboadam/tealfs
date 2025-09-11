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

type BalanceReqId string

type BalanceReq struct {
	Caller       model.NodeId
	BalanceReqId BalanceReqId
}

func (a *BalanceReq) Type() model.PayloadType {
	return model.AllBlockIdReqType
}

type BlockIdList struct {
	Caller       model.NodeId
	BlockIds     set.Set[model.BlockId]
	BalanceReqId BalanceReqId
}

func (a *BlockIdList) Type() model.PayloadType {
	return model.AllBlockIdRespType
}

type StoreItCmd struct {
	StoreItId StoreItId
	BlockId   model.BlockId
	Caller    model.NodeId
}

func (s *StoreItCmd) Type() model.PayloadType {
	return model.StoreItCmdType
}

type StoreItResp struct {
	StoreItId StoreItId
	Caller    model.NodeId
	Ok        bool
	Msg       string
}

func (s *StoreItResp) Type() model.PayloadType {
	return model.StoreItRespType
}

type StoreItId string
