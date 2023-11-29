package mgr

import (
	"tealfs/pkg/conns"
	"tealfs/pkg/proto"
	"tealfs/pkg/util"
)

func remoteIsMissingNodes(connlist conns.Conns, syncNodes *proto.SyncNodes) bool {
	localNodes := connlist.GetConns()
	remoteNodes := syncNodes.GetIds()

	if remoteNodes.Len() < localNodes.Len() {
		return true
	}

	for _, localId := range localNodes.GetValues() {
		if !remoteNodes.Exists(localId) {
			return true
		}
	}

	return false
}

func findMyMissingConns(connlist conns.Conns, syncNodes *proto.SyncNodes) *util.Set[conns.Conn] {
	result := util.NewSet[conns.Conn]()

	localNodes := connlist.GetConns()
	remoteNodes := syncNodes.GetIds()

	missingIds := remoteNodes.Minus(&localNodes)

	for _, missingId := range missingIds.GetValues() {
		node, ok := syncNodes.NodeForId(missingId)
		if ok {
			result.Add(conns.NewConn(node.Address))
		}
	}

	return &result
}
