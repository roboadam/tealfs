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

package tnet

import (
	"encoding/gob"
	"fmt"
	"io"
	"tealfs/pkg/blockreader"
	"tealfs/pkg/blocksaver"
	"tealfs/pkg/model"
	"tealfs/pkg/webdav"

	log "github.com/sirupsen/logrus"
)

type RawNet struct {
	decoder *gob.Decoder
	encoder *gob.Encoder
	conn    io.ReadWriteCloser
}

func NewRawNet(conn io.ReadWriteCloser) *RawNet {
	return &RawNet{
		decoder: gob.NewDecoder(conn),
		encoder: gob.NewEncoder(conn),
		conn:    conn,
	}
}

func (r *RawNet) Close() error {
	return r.conn.Close()
}

func (r *RawNet) ReadPayload() (model.Payload, error) {
	var payloadType model.PayloadType
	err := r.decoder.Decode(&payloadType)
	if err != nil {
		log.Error("failed to decode payload type: " + err.Error())
		return nil, err
	}

	switch payloadType {
	case model.IAmType:
		var payload model.IAm
		err = r.decoder.Decode(&payload)
		return &payload, err
	case model.WriteRequestType:
		var payload model.WriteRequest
		err = r.decoder.Decode(&payload)
		return &payload, err
	case model.ReadRequestType:
		var payload model.ReadRequest
		err = r.decoder.Decode(&payload)
		return &payload, err
	case model.ReadResultType:
		var payload model.ReadResult
		err = r.decoder.Decode(&payload)
		return &payload, err
	case model.FileBroadcastType:
		var payload webdav.FileBroadcast
		err = r.decoder.Decode(&payload)
		return &payload, err
	case model.AddDiskMsgType:
		var payload model.AddDiskMsg
		err = r.decoder.Decode(&payload)
		return &payload, err
	case model.DiskAddedMsgType:
		var payload model.DiskAddedMsg
		err = r.decoder.Decode(&payload)
		return &payload, err
	case model.SyncType:
		var payload model.SyncNodes
		err = r.decoder.Decode(&payload)
		return &payload, err
	case model.WriteResultType:
		var payload model.WriteResult
		err = r.decoder.Decode(&payload)
		return &payload, err
	case model.SaveToDiskReq:
		var payload blocksaver.SaveToDiskReq
		err := r.decoder.Decode(&payload)
		return &payload, err
	case model.SaveToDiskResp:
		var payload blocksaver.SaveToDiskResp
		err := r.decoder.Decode(&payload)
		return &payload, err
	case model.GetFromDiskReq:
		var payload blockreader.GetFromDiskReq
		err := r.decoder.Decode(&payload)
		return &payload, err
	case model.GetFromDiskResp:
		var payload blockreader.GetFromDiskResp
		err := r.decoder.Decode(&payload)
		return &payload, err
	}

	panic("unknown payload type: " + fmt.Sprint(payloadType))
}

func (r *RawNet) SendPayload(payload model.Payload) error {
	err := r.encoder.Encode(payload.Type())
	if err != nil {
		return err
	}
	err = r.encoder.Encode(payload)
	if err != nil {
		return err
	}
	return nil
}
