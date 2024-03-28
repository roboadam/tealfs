package mgr

import (
	"tealfs/pkg/nodes"
	"tealfs/pkg/set"
)

func remoteIsMissingNodes(local *set.Set[nodes.NodeNew], remote *set.Set[nodes.NodeNew]) bool {
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
