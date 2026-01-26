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
	Data map[K]bool
	mux  *sync.RWMutex
}

func NewSet[K comparable]() Set[K] {
	return Set[K]{Data: make(map[K]bool), mux: &sync.RWMutex{}}
}

func NewSetFromSlice[K comparable](input []K) Set[K] {
	result := NewSet[K]()

	for _, item := range input {
		result.Add(item)
	}

	return result
}

func NewSetFromMapKeys[K comparable, J any](input map[K]J) Set[K] {
	result := NewSet[K]()

	for k := range input {
		result.Add(k)
	}

	return result
}

func (s *Set[K]) initMux() {
	if s.mux == nil {
		s.mux = &sync.RWMutex{}
	}
}

func (s *Set[K]) Add(item K) bool {
	s.initMux()
	s.mux.Lock()
	defer s.mux.Unlock()
	if _, ok := s.Data[item]; ok {
		return false
	}
	s.Data[item] = true
	return true
}

func (s *Set[K]) AddAll(other *Set[K]) {
	s.initMux()
	s.mux.Lock()
	defer s.mux.Unlock()
	for _, item := range other.GetValues() {
		s.Data[item] = true
	}
}

func (s *Set[K]) Remove(item K) {
	s.initMux()
	s.mux.Lock()
	defer s.mux.Unlock()
	delete(s.Data, item)
}

func (s *Set[K]) Equal(b *Set[K]) bool {
	s.initMux()
	s.mux.RLock()
	defer s.mux.RUnlock()
	if len(s.Data) != len(b.Data) {
		return false
	}

	for key := range s.Data {
		if !b.contains(key) {
			return false
		}
	}

	return true
}

func (s *Set[K]) Contains(k K) bool {
	s.initMux()
	s.mux.RLock()
	defer s.mux.RUnlock()
	return s.contains(k)
}

func (s *Set[K]) contains(k K) bool {
	_, exists := s.Data[k]
	return exists
}

func (s *Set[K]) GetValues() []K {
	s.initMux()
	s.mux.RLock()
	defer s.mux.RUnlock()
	result := make([]K, len(s.Data))
	i := 0
	for k := range s.Data {
		result[i] = k
		i++
	}
	return result
}

func (s *Set[K]) Minus(o *Set[K]) *Set[K] {
	s.initMux()
	s.mux.RLock()
	defer s.mux.RUnlock()
	result := NewSet[K]()
	for k := range s.Data {
		if !o.Contains(k) {
			result.Add(k)
		}
	}
	return &result
}

func (s *Set[K]) Pop() (K, *Set[K], bool) {
	clone := s.Clone()
	if clone.Len() == 0 {
		return *new(K), &clone, false
	}

	raw := clone.GetValues()
	k := raw[0]
	remainder := NewSetFromSlice(raw[1:])
	return k, &remainder, true
}

func (s *Set[K]) Len() int {
	s.initMux()
	s.mux.RLock()
	defer s.mux.RUnlock()
	return len(s.Data)
}

func (s *Set[K]) Clone() Set[K] {
	s.initMux()
	s.mux.RLock()
	defer s.mux.RUnlock()
	return Set[K]{
		Data: maps.Clone(s.Data),
		mux:  &sync.RWMutex{},
	}
}
