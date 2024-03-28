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
