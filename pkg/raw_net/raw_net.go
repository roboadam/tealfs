package raw_net

import (
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

func StringTo(conn net.Conn, value string) error {
	networkValue := []byte(value)
	_, err := conn.Write(networkValue)
	return err
}
