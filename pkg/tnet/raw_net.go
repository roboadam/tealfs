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

package tnet

import (
	"encoding/gob"
	"fmt"
	"io"
	"net"
	"tealfs/pkg/model"

	log "github.com/sirupsen/logrus"
)

func ReadPayload(conn io.Reader) (model.Payload, error) {
	decoder := gob.NewDecoder(conn)
	var payloadType model.PayloadType
	err := decoder.Decode(&payloadType)
	if err != nil {
		log.Error("failed to decode payload type: " + err.Error())
		return nil, err
	} else {
		fmt.Println("Received payload of type:", payloadType)
	}

	switch payloadType {
	case model.IAmType:
		var payload model.IAm
		err = decoder.Decode(&payload)
		if err != nil {
			log.Error("failed to decode IAm: " + err.Error())
		}
		return &payload, err
	case model.WriteRequestType:
		var payload model.WriteRequest
		err = decoder.Decode(&payload)
		if err != nil {
			log.Error("failed to decode WriteRequest: " + err.Error())
		}
		return &payload, err
	case model.ReadRequestType:
		var payload model.ReadRequest
		err = decoder.Decode(&payload)
		if err != nil {
			log.Error("failed to decode ReadRequest: " + err.Error())
		}
		return &payload, err
	case model.ReadResultType:
		var payload model.ReadResult
		err = decoder.Decode(&payload)
		if err != nil {
			log.Error("failed to decode ReadResult: " + err.Error())
		}
		return &payload, err
	case model.BroadcastType:
		var payload model.Broadcast
		err = decoder.Decode(&payload)
		if err != nil {
			log.Error("failed to decode Broadcast: " + err.Error())
		}
		return &payload, err
	case model.AddDiskRequestType:
		var payload model.AddDiskReq
		err = decoder.Decode(&payload)
		if err != nil {
			log.Error("failed to decode AddDiskReq: " + err.Error())
		}
		return &payload, err
	case model.SyncType:
		var payload model.SyncNodes
		err = decoder.Decode(&payload)
		if err != nil {
			log.Error("failed to decode SyncNodes: " + err.Error())
		}
		return &payload, err
	case model.WriteResultType:
		var payload model.WriteResult
		err = decoder.Decode(&payload)
		if err != nil {
			log.Error("failed to decode WriteResult: " + err.Error())
		}
		return &payload, err
	}

	panic("unknown payload type: " + fmt.Sprint(payloadType))
}

func SendPayload(conn io.Writer, payload model.Payload) error {
	encoder := gob.NewEncoder(conn)
	err := encoder.Encode(payload.Type())
	if err != nil {
		panic("failed to encode payload type: " + err.Error())
		// return err
	}
	fmt.Println("Sending payload of type:", payload.Type())
	err = encoder.Encode(payload)
	if err != nil {
		panic("failed to encode payload: " + err.Error())
		// return err
	}
	return nil
}

func ReadBytes(conn net.Conn, length uint32) ([]byte, error) {
	buf := make([]byte, length)
	offset := uint32(0)

	for offset < length {
		numBytes, err := conn.Read(buf[offset:])
		if err != nil {
			return nil, err
		}
		offset += uint32(numBytes)
	}

	return buf, nil
}

func SendBytes(conn net.Conn, data []byte) error {
	bytesWritten := 0
	for bytesWritten < len(data) {
		numBytes, err := conn.Write(data[bytesWritten:])
		if err != nil {
			return err
		}
		bytesWritten += numBytes
	}
	return nil
}
