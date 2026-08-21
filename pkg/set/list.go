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

type List[K any] struct {
	Data []K
	mux  *sync.RWMutex
}

func NewList[K any]() List[K] {
	return List[K]{Data: make([]K, 1), mux: &sync.RWMutex{}}
}

func NewListFromSlice[K any](input []K) List[K] {
	result := NewList[K]()

	for _, item := range input {
		result.Add(item)
	}

	return result
}

func (s *List[K]) initMux() {
	if s.mux == nil {
		s.mux = &sync.RWMutex{}
	}
}

func (s *List[K]) Add(item K) {
	s.initMux()
	s.mux.Lock()
	defer s.mux.Unlock()
	s.Data = append(s.Data, item)
}

func (s *List[K]) AddAll(other *List[K]) {
	s.initMux()
	s.mux.Lock()
	defer s.mux.Unlock()
	s.Data = append(s.Data, other.GetValues()...)
}

func (s *List[K]) GetValues() []K {
	s.initMux()
	s.mux.RLock()
	defer s.mux.RUnlock()
	result := make([]K, len(s.Data))
	i := 0
	for _, k := range s.Data {
		result[i] = k
		i++
	}
	return result
}

func (s *List[K]) Len() int {
	s.initMux()
	s.mux.RLock()
	defer s.mux.RUnlock()
	return len(s.Data)
}

func (s *List[K]) Clone() List[K] {
	s.initMux()
	s.mux.RLock()
	defer s.mux.RUnlock()
	newSlice := make([]K, len(s.Data))
	copy(newSlice, s.Data)
	return List[K]{
		Data: newSlice,
		mux:  &sync.RWMutex{},
	}
}
