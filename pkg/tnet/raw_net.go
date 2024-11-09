// Copyright (C) 2024 Adam Hess
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
	"fmt"
	"net"
)

func ReadPayload(conn net.Conn) ([]byte, error) {
	rawLen, err := ReadBytes(conn, 4)
	if err != nil {
		return nil, err
	}
	fmt.Println("rawnet read rawlen:", rawLen)
	size := binary.BigEndian.Uint32(rawLen)
	a, b := ReadBytes(conn, size)
	fmt.Println("rawnet read rawbytes:", a)
	return a, b
}

func SendPayload(conn net.Conn, data []byte) error {
	size := uint32(len(data))
	buf := make([]byte, 4)
	binary.BigEndian.PutUint32(buf, size)
	err := SendBytes(conn, buf)
	if err != nil {
		return err
	}
	fmt.Println("rawnet send rawlen:", buf)
	err = SendBytes(conn, data)
	if err != nil {
		return err
	}
	fmt.Println("rawnet send rawbytes:", data)
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
