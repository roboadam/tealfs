package proto

import (
	"bytes"
	"encoding/binary"
	hash2 "tealfs/pkg/hash"
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

func BytesFromBytes(data []byte) ([]byte, []byte) {
	length, remainder := IntFromBytes(data)
	result := remainder[:length]
	return result, remainder[length:]
}

func BytesToBytes(value []byte) []byte {
	length := uint32(len(value))
	rawLength := IntToBytes(length)
	return append(rawLength, value...)
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

func BoolFromBytes(value []byte) (bool, []byte) {
	return value[0] == 1, value[1:]
}

func BlockToBytes(value store.Block) []byte {
	id := StringToBytes(string(value.Id))
	data := BytesToBytes(value.Data)
	hash := BytesToBytes(value.Hash.Value)
	numChildren := IntToBytes(uint32(len(value.Children)))

	result := bytes.Join([][]byte{id, data, hash, numChildren}, []byte{})
	for _, child := range value.Children {
		result = append(result, StringToBytes(string(child))...)
	}

	return result
}

func BlockFromBytes(value []byte) (store.Block, []byte) {
	id, remainder := StringFromBytes(value)
	data, remainder := BytesFromBytes(remainder)
	hash, remainder := BytesFromBytes(remainder)
	numChildren, remainder := IntFromBytes(remainder)

	var children []store.Id
	for numChildren > 0 {
		numChildren--
		var child string
		child, remainder = StringFromBytes(remainder)
		children = append(children, store.Id(child))
	}
	return store.Block{
		Id:       store.Id(id),
		Data:     data,
		Hash:     hash2.FromRaw(hash),
		Children: children,
	}, remainder
}

func AddType(id uint8, data []byte) []byte {
	buf := make([]byte, len(data)+1)
	buf[0] = id
	copy(buf[1:], data)
	return buf
}
