package mgr

import (
	"tealfs/pkg/conns"
	"tealfs/pkg/node"
	"tealfs/pkg/proto"
	"tealfs/pkg/util"
)

func missingConns(connlist conns.Conns, syncNodes proto.SyncNodes) *util.Set[node.Id] {
	localNodes := connlist.GetConns()
	remoteNodes := syncNodes.GetIds()
	return localNodes.Minus(&remoteNodes)
}
