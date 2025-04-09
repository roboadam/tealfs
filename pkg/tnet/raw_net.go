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
	"context"
	"encoding/binary"
	"net"
)

func ReadPayload(ctx context.Context, conn net.Conn) ([]byte, error) {
	rawLen, err := ReadBytes(ctx, conn, 4)
	if err != nil {
		return nil, err
	}
	size := binary.BigEndian.Uint32(rawLen)
	a, b := ReadBytes(ctx, conn, size)
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
	err = SendBytes(conn, data)
	if err != nil {
		return err
	}
	return nil
}

func ReadBytes(ctx context.Context, conn net.Conn, length uint32) ([]byte, error) {
	buf := make([]byte, length)
	offset := uint32(0)

	for offset < length {
		numBytes, err := readWithContext(ctx, conn, buf[offset:])
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

func readWithContext(ctx context.Context, conn net.Conn, buf []byte) (int, error) {
	done := make(chan struct{})
	defer close(done)

	go func() {
		select {
		case <-ctx.Done():
			conn.Close()
			// conn.SetReadDeadline(time.Now())
		case <-done:
		}
	}()

	n, err := conn.Read(buf)
	if ctx.Err() != nil {
		return n, ctx.Err()
	}

	return n, err
}
