// Copyright (C) 2024 Adam Hess
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
	"tealfs/pkg/nodes"
	"tealfs/pkg/set"
)

func remoteIsMissingNodes(local *set.Set[nodes.Node], remote *set.Set[nodes.Node]) bool {
	if remote.Len() < local.Len() {
		return true
	}

	for _, localId := range local.GetValues() {
		if !remote.Exists(localId) {
			return true
		}
	}

	return false
}
