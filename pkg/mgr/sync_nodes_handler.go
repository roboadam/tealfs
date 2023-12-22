package mgr

import (
	"tealfs/pkg/model/node"
	"tealfs/pkg/proto"
	"tealfs/pkg/tnet"
	"tealfs/pkg/util"
)

func remoteIsMissingNodes(connlist tnet.Conns, syncNodes *proto.SyncNodes) bool {
	localNodes := connlist.GetIds()
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

func findMyMissingConns(connlist tnet.Conns, syncNodes *proto.SyncNodes) *util.Set[tnet.Conn] {
	result := util.NewSet[tnet.Conn]()

	localNodes := connlist.GetIds()
	remoteNodes := syncNodes.GetIds()
	myNode := util.NewSet[node.Id]()
	myNode.Add(connlist.MyNodeId)

	missingIds := remoteNodes.Minus(&localNodes).Minus(&myNode)

	for _, missingId := range missingIds.GetValues() {
		n, ok := syncNodes.NodeForId(missingId)
		if ok {
			result.Add(tnet.NewConn(n.Address))
		}
	}

	return &result
}
