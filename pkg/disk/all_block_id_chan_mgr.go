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

package disk

import (
	"tealfs/pkg/model"
	"tealfs/pkg/rebalancer"
	"tealfs/pkg/set"
)

type AllBlockIdChanMgr struct {
	ReqChans  set.Bimap[model.DiskId, chan rebalancer.AllBlockIdReq]
	RespChans set.Bimap[model.DiskId, chan rebalancer.AllBlockIdResp]
}

func (a *AllBlockIdChanMgr) Add(
	diskId model.DiskId,
	reqChan chan rebalancer.AllBlockIdReq,
	respChan chan rebalancer.AllBlockIdResp,
) {
	a.ReqChans.Add(diskId, reqChan)
	a.RespChans.Add(diskId, respChan)
}
