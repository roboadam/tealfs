package raw_net

import (
	"encoding/binary"
	"errors"
	"io"
	"net"
	"tealfs/pkg/proto"
)

func CommandAndLength(conn net.Conn) (proto.NetCmd, uint32, error) {
	buf := make([]byte, 9)
	_, err := conn.Read(buf)
	if err != nil {
		return proto.NetCmd{Value: 0}, 0, err
	}

	return proto.netCmd{Value: buf[0]}, 0, nil
}

func ReadBytes(conn net.Conn, length uint32) ([]byte, error) {
	buf := make([]byte, length)
	totalBytes := uint32(0)

	for totalBytes < length {
		numBytes, err := conn.Read(buf)
		if err != nil {
			return nil, err
		}
		totalBytes += uint32(numBytes)
	}

	return buf, nil
}

func Int8From(conn net.Conn) (uint8, error) {
	buf := make([]byte, 1)
	_, err := conn.Read(buf)
	if err != nil {
		return 0, err
	}

	return buf[0], nil
}

func UInt32From(conn net.Conn) (uint32, error) {
	buf := make([]byte, 4)
	_, err := conn.Read(buf)
	if err != nil {
		return 0, err
	}

	value := binary.BigEndian.Uint32(buf)

	return value, nil
}

func StringFrom(conn net.Conn, length int) (string, error) {
	buffer := make([]byte, length)
	bytesRead := 0

	for bytesRead < length {
		n, err := conn.Read(buffer[bytesRead:])
		if err != nil {
			if err == io.EOF {
				return "", errors.New("EOF reached before reading the specified number of bytes")
			}
			return "", err
		}
		bytesRead += n
	}

	utf8String := string(buffer)
	return utf8String, nil
}

func Int8To(conn net.Conn, value int8) error {
	networkValue := []byte{byte(value)}
	_, err := conn.Write(networkValue)
	return err
}

func UInt32To(conn net.Conn, value uint32) error {
	networkValue := make([]byte, 4)
	binary.BigEndian.PutUint32(networkValue, value)
	_, err := conn.Write(networkValue)
	return err
}

func StringTo(conn net.Conn, value string) error {
	networkValue := []byte(value)
	_, err := conn.Write(networkValue)
	return err
}
