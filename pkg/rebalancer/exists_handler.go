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

import "context"

type ExistsHandler struct {
	InExistsReq   <-chan ExistsReq
	OutExistsResp chan<- ExistsResp
}

func (e *ExistsHandler) Start(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			return
		case req := <-e.InExistsReq:
			e.handleExistsReq(req)
		}
	}
}

func (e *ExistsHandler) handleExistsReq(req ExistsReq) {
	// TODO: Check if the block actually exists on the disk
	e.OutExistsResp <- ExistsResp{
		Req: req,
		Ok:  true,
		Msg: "Block exists",
	}
}
