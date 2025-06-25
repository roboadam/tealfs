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
	"encoding/binary"
	"net"
)

func ReadPayload(conn net.Conn) (byte, []byte, error) {
	rawLen, err := ReadBytes(conn, 4)
	if err != nil {
		return 0xff, nil, err
	}
	size := binary.BigEndian.Uint32(rawLen)
	bytes, err := ReadBytes(conn, size)
	if err != nil {
		return 0xff, nil, err
	}
	if len(bytes) > 0 {
		return bytes[0], bytes[1:], nil
	}
	return 0xff, bytes, err
}

func SendPayload(conn net.Conn, data []byte) error {
	typ := []byte{0}
	dataWithType := append(typ, data...)
	size := uint32(len(dataWithType))
	buf := make([]byte, 4)
	binary.BigEndian.PutUint32(buf, size)
	err := SendBytes(conn, buf)
	if err != nil {
		return err
	}
	err = SendBytes(conn, dataWithType)
	if err != nil {
		return err
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
