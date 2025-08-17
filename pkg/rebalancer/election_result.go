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
	"context"
	"time"
)

type ElectionResult struct {
	InCollectResult <-chan struct{}
	InAlive         <-chan Alive

	end *time.Time
}

const TIMEOUT = time.Second * 30

func (e *ElectionResult) Start(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			return
		case <-e.InCollectResult:
			end := time.Now()
			end.Add(TIMEOUT)
		case a := <-e.InAlive:
			if e.end == nil {
				continue
			}
			if time.Now().After(*e.end) { 
				e.end = nil
			}
			
		}
	}
}
