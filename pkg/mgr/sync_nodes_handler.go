package mgr

import (
	"tealfs/pkg/conns"
	"tealfs/pkg/node"
	"tealfs/pkg/proto"
	"tealfs/pkg/util"
)

func findMissingConns(connlist conns.Conns, syncNodes *proto.SyncNodes) *util.Set[node.Node] {
	localNodes := connlist.GetConns()
	remoteNodes := syncNodes.GetIds()
	result := util.NewSet[node.Node]()
	missingIds := localNodes.Minus(&remoteNodes)
	for missingId := range missingIds.GetValues() {
		su
	}
}
