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
