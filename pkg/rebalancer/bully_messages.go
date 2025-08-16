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

import "tealfs/pkg/model"

type Election struct {
	NodeID model.NodeId
}

func (e *Election) Type() model.PayloadType {
	return model.Election
}

type Alive struct {
	NodeID model.NodeId
}

func (a *Alive) Type() model.PayloadType {
	return model.Alive
}

type Victory struct {
	NodeID model.NodeId
}

func (v *Victory) Type() model.PayloadType {
	return model.Victory
}
