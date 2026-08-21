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

package set

import (
	"sync"
)

type Counter[K comparable] struct {
	Data map[K]int
	mux  *sync.RWMutex
}

func (c *Counter[K]) init() {
	if c.Data == nil {
		c.Data = make(map[K]int)
	}
	if c.mux == nil {
		c.mux = &sync.RWMutex{}
	}
}

func (c *Counter[K]) Tick(key K) {
	c.init()
	c.mux.Lock()
	defer c.mux.Unlock()
	if _, ok := c.Data[key]; !ok {
		c.Data[key] = 0
	}
	c.Data[key]++
}

func (c *Counter[K]) Count(key K) int {
	c.init()
	c.mux.RLock()
	defer c.mux.RUnlock()
	if _, ok := c.Data[key]; !ok {
		return 0
	}
	return c.Data[key]
}
