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

package conns

import (
	"context"
	"tealfs/pkg/model"
)

type Reconnector struct {
	OutConnectTo chan<- model.ConnectToNodeReq

	Mapper *model.NodeConnectionMapper
}

func (r *Reconnector) Start(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			return
		default:
			r.reConnect()
		}
	}
}

func (r *Reconnector) reConnect() {
	addresses := r.Mapper.AddressesWithoutConnections()
	for _, address := range addresses.GetValues() {
		r.OutConnectTo <- model.ConnectToNodeReq{
			Address: address,
		}
	}
}
