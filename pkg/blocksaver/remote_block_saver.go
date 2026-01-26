// Copyright (C) 2026 Adam Hess
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

package blocksaver

import (
	"context"
	"errors"
	"tealfs/pkg/model"
)

type RemoteBlockSaver struct {
	Req         <-chan SaveToDiskReq
	Sends       chan<- model.SendPayloadMsg
	NoConnResp  chan<- SaveToDiskResp
	NodeConnMap *model.NodeConnectionMapper
}

func (rbs *RemoteBlockSaver) Start(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			return
		case req := <-rbs.Req:
			// Send the payload to the remote node if it has a connection, otherwise send back an error response
			conn, ok := rbs.NodeConnMap.ConnForNode(req.Dest.NodeId)
			if ok {
				rbs.Sends <- model.SendPayloadMsg{
					ConnId:  conn,
					Payload: &req,
				}
			} else {
				rbs.NoConnResp <- SaveToDiskResp{
					Caller: req.Caller,
					Dest:   req.Dest,
					Resp: model.PutBlockResp{
						Id:  req.Req.Id,
						Err: errors.New("No connection to node " + string(req.Dest.NodeId)),
					},
				}
			}
		}
	}
}
