package raw_net

import (
	"encoding/binary"
	"net"
)

func IntFrom(conn net.Conn) (uint32, error) {
	buf := make([]byte, 4)
	_, err := conn.Read(buf)
	if err != nil {
		return 0, err
	}

	intValue := binary.BigEndian.Uint32(buf)
	return intValue, nil
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
