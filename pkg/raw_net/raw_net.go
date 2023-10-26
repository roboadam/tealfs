package raw_net

import (
	"encoding/binary"
	"errors"
	"io"
	"net"
)

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
