// Copyright (C) 2025 Adam Hess
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

package blockreader

import (
	"context"
	"errors"
	"tealfs/pkg/model"
)

type RemoteBlockReader struct {
	Req         <-chan GetFromDiskReq
	Sends       chan<- model.SendPayloadMsg
	NoConnResp  chan<- GetFromDiskResp
	NodeConnMap *model.NodeConnectionMapper
}

func (rbs *RemoteBlockReader) Start(ctx context.Context) {
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
				rbs.NoConnResp <- GetFromDiskResp{
					Caller: req.Caller,
					Dest:   req.Dest,
					Resp: model.GetBlockResp{
						Id:  req.Req.Id,
						Err: errors.New("no connection to get from that node"),
					},
				}
			}
		}
	}
}
