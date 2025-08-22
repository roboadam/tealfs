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

package set

import (
	"maps"
	"sync"
)

type Map[K comparable, J any] struct {
	Data map[K]J
	mux  *sync.RWMutex
}

func NewMap[K comparable, J any]() Map[K, J] {
	return Map[K, J]{Data: make(map[K]J), mux: &sync.RWMutex{}}
}

func (s *Map[K, J]) initMux() {
	if s.mux == nil {
		s.mux = &sync.RWMutex{}
	}
}

func (s *Map[K, J]) Add(key K, item J) {
	s.initMux()
	s.mux.Lock()
	defer s.mux.Unlock()
	s.Data[key] = item
}

func (s *Map[K, J]) Get(key K) (J, bool) {
	s.initMux()
	s.mux.RLock()
	defer s.mux.RUnlock()
	value, ok := s.Data[key]
	return value, ok
}

func (s *Map[K, J]) Remove(key K) {
	s.initMux()
	s.mux.Lock()
	defer s.mux.Unlock()
	delete(s.Data, key)
}

func (s *Map[K, J]) Len() int {
	s.initMux()
	s.mux.RLock()
	defer s.mux.RUnlock()
	return len(s.Data)
}

func (s *Map[K, J]) Clone() Map[K, J] {
	s.initMux()
	s.mux.RLock()
	defer s.mux.RUnlock()
	return Map[K, J]{
		Data: maps.Clone(s.Data),
		mux:  &sync.RWMutex{},
	}
}
