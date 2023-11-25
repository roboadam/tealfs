package mgr

import (
	"fmt"
	"tealfs/pkg/conns"
	"tealfs/pkg/proto"
)

func missingConns(connlist conns.Conns, syncNodes proto.SyncNodes) []conns.Conn {
	localNodes := connlist.GetConns()
	remoteNodes := syncNodes.Nodes
	fmt.Println(localNodes, remoteNodes)
	result := make([]conns.Conn, 0)
	return result
}
