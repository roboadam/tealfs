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

package mgr

import (
	"tealfs/pkg/model"
	"tealfs/pkg/set"
)

type pendingBlockWrites struct {
	b2ptr map[model.BlockId]set.Set[model.DiskPointer]
	ptr2b map[model.DiskPointer]model.BlockId
}

func newPendingBlockWrites() pendingBlockWrites {
	return pendingBlockWrites{
		b2ptr: make(map[model.BlockId]set.Set[model.DiskPointer]),
		ptr2b: make(map[model.DiskPointer]model.BlockId),
	}
}

func (p *pendingBlockWrites) add(b model.BlockId, ptr model.DiskPointer) {
	if _, exists := p.b2ptr[b]; !exists {
		p.b2ptr[b] = set.NewSet[model.DiskPointer]()
	}

	s := p.b2ptr[b]
	s.Add(ptr)

	p.ptr2b[ptr] = b
}

type resolveResult int

const (
	done resolveResult = iota
	notDone
	notTracking
)

func (p *pendingBlockWrites) resolve(ptr model.DiskPointer) (resolveResult, model.BlockId) {
	if b, exists := p.ptr2b[ptr]; exists {
		s := p.b2ptr[b]
		s.Remove(ptr)
		delete(p.ptr2b, ptr)
		if s.Len() == 0 {
			delete(p.b2ptr, b)
			return done, b
		}
		return notDone, b
	}
	return notTracking, ""
}

func (p *pendingBlockWrites) cancel(b model.BlockId) {
	if s, exists := p.b2ptr[b]; exists {
		for _, ptr := range s.GetValues() {
			delete(p.ptr2b, ptr)
		}
		delete(p.b2ptr, b)
	}
}
