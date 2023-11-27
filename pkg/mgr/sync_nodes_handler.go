package mgr

import (
	"tealfs/pkg/conns"
	"tealfs/pkg/proto"
	"tealfs/pkg/util"
)

func findMissingConns(connlist conns.Conns, syncNodes *proto.SyncNodes) *util.Set[conns.Conn] {
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
