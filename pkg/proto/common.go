package proto

import (
	"encoding/binary"
)

const CommandAndLengthSize uint32 = 9

func StringFromBytes(data []byte) (string, []byte) {
	length, data := IntFromBytes(data)
	utfString := string(data[:length])
	return utfString, data[length:]
}

func StringToBytes(value string) []byte {
	rawString := []byte(value)
	length := uint32(len(rawString))
	rawLength := IntToBytes(length)
	return append(rawLength, rawString...)
}

func IntFromBytes(data []byte) (uint32, []byte) {
	value := binary.BigEndian.Uint32(data)
	return value, data[4:]
}

func IntToBytes(value uint32) []byte {
	buf := make([]byte, 4)
	binary.BigEndian.PutUint32(buf, value)
	return buf
}

func CommandAndLengthFromBytes(data []byte) (NetCmd, uint32, []byte) {
	length := binary.BigEndian.Uint32(data[1:9])
	return NetCmd{Value: data[0]}, length, data[9:]
}

func CommandAndLengthToBytes(cmd NetCmd, length uint32) []byte {
	buf := make([]byte, 9)
	buf[0] = cmd.Value
	binary.BigEndian.PutUint32(buf[1:9], length)
	return buf
}
