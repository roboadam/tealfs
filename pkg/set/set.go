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

type Set[K comparable] struct {
	mux  sync.RWMutex
	data map[K]bool
}

func NewSet[K comparable]() Set[K] {
	return Set[K]{data: make(map[K]bool), mux: sync.RWMutex{}}
}

func NewSetFromMapKeys[K comparable, J any](input map[K]J) Set[K] {
	result := Set[K]{data: make(map[K]bool), mux: sync.RWMutex{}}

	for k := range input {
		result.Add(k)
	}

	return result
}

func (s *Set[K]) Add(item K) {
	s.mux.Lock()
	defer s.mux.Unlock()
	s.data[item] = true
}

func (s *Set[K]) Remove(item K) {
	s.mux.Lock()
	defer s.mux.Unlock()
	delete(s.data, item)
}

func (s *Set[K]) Equal(b *Set[K]) bool {
	s.mux.RLock()
	defer s.mux.RUnlock()
	if len(s.data) != len(b.data) {
		return false
	}

	for key := range s.data {
		if !b.Contains(key) {
			return false
		}
	}

	return true
}

func (s *Set[K]) Contains(k K) bool {
	s.mux.RLock()
	defer s.mux.RUnlock()
	_, exists := s.data[k]
	return exists
}

func (s *Set[K]) GetValues() []K {
	s.mux.RLock()
	defer s.mux.RUnlock()
	result := make([]K, len(s.data))
	i := 0
	for k := range s.data {
		result[i] = k
		i++
	}
	return result
}

func (s *Set[K]) Minus(o *Set[K]) *Set[K] {
	s.mux.RLock()
	defer s.mux.RUnlock()
	result := NewSet[K]()
	for k := range s.data {
		if !o.Contains(k) {
			result.Add(k)
		}
	}
	return &result
}

func (s *Set[K]) Len() int {
	s.mux.RLock()
	defer s.mux.RUnlock()
	return len(s.data)
}

func (s *Set[K]) Clone() Set[K] {
	s.mux.RLock()
	defer s.mux.RUnlock()
	return Set[K]{
		data: maps.Clone(s.data),
	}
}
