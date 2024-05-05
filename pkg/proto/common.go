package proto

import (
	"encoding/binary"
	"tealfs/pkg/store"
)

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

func BoolToBytes(value bool) []byte {
	result := []byte{1}
	if value {
		result[0] = 0
	} else {
		result[0] = 1
	}
	return result
}

func BlockToBytes(value store.Block) []byte {
	result := make([]byte, 0)
	id := StringToBytes(string(value.Id))
	data := value.Data
	hash := value.Hash.Value

	result = append(result, id...)
	result = append(result, data...)
	result = append(result, hash...)
	for _, child := range value.Children {
		result = append(result, StringToBytes(string(child))...)
	}
	return result
}

func AddType(id uint8, data []byte) []byte {
	buf := make([]byte, len(data)+1)
	buf[0] = id
	copy(buf[1:], data)
	return buf
}
