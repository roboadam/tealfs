package mgr

import (
	"tealfs/pkg/model/node"
	"tealfs/pkg/proto"
	"tealfs/pkg/set"
	"tealfs/pkg/tnet"
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

func findMyMissingConns(connlist tnet.Conns, syncNodes *proto.SyncNodes) *set.Set[tnet.Conn] {
	result := set.NewSet[tnet.Conn]()

	localNodes := connlist.GetIds()
	remoteNodes := syncNodes.GetIds()
	myNode := set.NewSet[node.Id]()
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
