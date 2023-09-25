package raw_net

import (
	"encoding/binary"
	"net"
)

func IntFrom(conn net.Conn) (uint32, error) {
	buf := make([]byte, 4) // 4 bytes for a 32-bit integer
	_, err := conn.Read(buf)
	if err != nil {
		return 0, err
	}

	// Convert the received bytes to an integer (big endian)
	intValue := binary.BigEndian.Uint32(buf)
	return intValue, nil
}

func IntTo(conn net.Conn, value uint32) error {
	buf := make([]byte, 4) // 4 bytes for a 32-bit integer

	// Convert the integer to bytes in network byte order (big endian)
	binary.BigEndian.PutUint32(buf, value)

	// Send the byte slice over the network connection
	_, err := conn.Write(buf)
	return err
}
