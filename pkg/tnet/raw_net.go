package tnet

import (
	"encoding/binary"
	"net"
)

func ReadPayload(conn net.Conn) ([]byte, error) {
	println("1 i'm", conn.LocalAddr().String(), "reading from", conn.RemoteAddr().String())
	rawLen, err := ReadBytes(conn, 4)
	if err != nil {
		return nil, err
	}
	println("2 i'm", conn.LocalAddr().String(), "reading from", conn.RemoteAddr().String())
	size := binary.BigEndian.Uint32(rawLen)
	a, b := ReadBytes(conn, size)
	println("3 i'm", conn.LocalAddr().String(), "reading from", conn.RemoteAddr().String())
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

func ReadBytes(conn net.Conn, length uint32) ([]byte, error) {
	buf := make([]byte, length)
	offset := uint32(0)

	for offset < length {
		numBytes, err := conn.Read(buf[offset:])
		if err != nil {
			return nil, err
		}
		offset += uint32(numBytes)
		//println("looping")
	}

	return buf, nil
}

func SendBytes(conn net.Conn, data []byte) error {
	println("sending ", data[0], "i'm", conn.LocalAddr().String(), "sending to", conn.RemoteAddr().String())
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
