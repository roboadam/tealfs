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

type ListOnDiskBlockIdsCmd struct {
	Caller       model.NodeId
	BalanceReqId BalanceReqId
}

func (a *ListOnDiskBlockIdsCmd) Type() model.PayloadType {
	return model.BalanceReqType
}

type OnDiskBlockIdList struct {
	Caller       model.NodeId
	BlockIds     set.Set[model.BlockId]
	BalanceReqId BalanceReqId
}

func (a *OnDiskBlockIdList) Type() model.PayloadType {
	return model.OnDiskBlockIdListType
}

type FilesystemBlockIdList struct {
	Caller       model.NodeId
	BlockIds     set.Set[model.BlockId]
	BalanceReqId BalanceReqId
}

func (f *FilesystemBlockIdList) Type() model.PayloadType {
	return model.FilesystemBlockIdListType
}

type StoreItReq struct {
	Caller       model.NodeId
	BalanceReqId BalanceReqId
	StoreItId    StoreItId
	DestNodeId   model.NodeId
	DestDiskId   model.DiskId
	DestBlockId  model.BlockId
}

type StoreItCmd struct {
	BalanceReqId BalanceReqId
	StoreItId    StoreItId
	BlockId      model.BlockId
	Caller       model.NodeId
	Recipient    model.NodeId
}

func (s *StoreItCmd) Type() model.PayloadType {
	return model.StoreItCmdType
}

type StoreItResp struct {
	Req StoreItReq
	Ok  bool
	Msg string
}

func (s *StoreItResp) Type() model.PayloadType {
	return model.StoreItRespType
}

type StoreItId string

type ExistsId string
type ExistsReq struct {
	Caller       model.NodeId
	BalanceReqId BalanceReqId
	ExistsId     ExistsId
	DestNodeId   model.NodeId
	DestDiskId   model.DiskId
	DestBlockId  model.BlockId
}

func (e *ExistsReq) Type() model.PayloadType {
	return model.ExistsReqType
}

type ExistsResp struct {
	Req ExistsReq
	Ok  bool
	Msg string
}

func (e *ExistsResp) Type() model.PayloadType {
	return model.ExistsRespType
}
