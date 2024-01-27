package tnet

import (
	"encoding/binary"
	"net"
)

func ReadPayload(conn net.Conn) ([]byte, error) {
	rawLen, err := ReadBytes(conn, 4)
	if err != nil {
		return nil, err
	}
	size := binary.BigEndian.Uint32(rawLen)
	return ReadBytes(conn, size)
}

func SendPayload(conn net.Conn, data []byte) error {
	size := uint32(len(data))
	buf := make([]byte, 4)
	binary.BigEndian.PutUint32(buf, size)
	println("raw_net.SendPayload1")
	err := SendBytes(conn, buf)
	if err != nil {
		return err
	}
	println("raw_net.SendPayload2")
	err = SendBytes(conn, data)
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
