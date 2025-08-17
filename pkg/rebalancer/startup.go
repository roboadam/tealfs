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
	"tealfs/pkg/model"
	"tealfs/pkg/set"
)

type Startup struct {
	NodeId model.NodeId
	Mapper *model.NodeConnectionMapper
}

func (s *Startup) Start(ctx context.Context) {
	greaterNodes := set.NewSet[model.NodeId]()

	nodes := s.Mapper.ConnectedNodes()
	for _, node := range nodes.GetValues() {
		if node > s.NodeId {
			greaterNodes.Add(node)
		}
	}

	
}
