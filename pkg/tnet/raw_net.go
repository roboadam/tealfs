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
	"net"
	"tealfs/pkg/model"
)

func ReadPayload(conn net.Conn) (model.Payload, error) {
	decoder := gob.NewDecoder(conn)
	var result model.Payload
	err := decoder.Decode(&result)
	return result, err
}

func SendPayload(conn net.Conn, payload *model.Payload) error {
	fmt.Printf("Type: %T\n", payload)
	fmt.Println("1")
	encoder := gob.NewEncoder(conn)
	fmt.Println("2")
	err := encoder.Encode(payload)
	fmt.Println("3")
	if err != nil {
		fmt.Println("4")
		return err
	}
	fmt.Println("5")
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
