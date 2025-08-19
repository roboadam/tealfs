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

type CollectAllBlockIds struct {
	InAllBlockIdResp <-chan AllBlockIdResp
	collector        map[AllBlockId][]AllBlockIdResp
}

func (c *CollectAllBlockIds) Start(ctx context.Context) {
	c.collector = make(map[AllBlockId][]AllBlockIdResp)
	for {
		select {
		case <-ctx.Done():
			return
		case resp := <-c.InAllBlockIdResp:
			c.collector[resp.Id] = append(c.collector[resp.Id], resp)
		}
	}
}
