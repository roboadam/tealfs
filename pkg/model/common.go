// Copyright (C) 2024 Adam Hess
//
// This program is free software: you can redistribute it and/or modify it under
// the terms of the GNU Affero General Public License as published by the Free
// Software Foundation, version 3.
//
// This program is distributed in the hope that it will be useful, but WITHOUT
// ANY WARRANTY; without even the implied warranty of MERCHANTABILITY or
// FITNESS FOR A PARTICULAR PURPOSE. See the GNU Affero General Public License
// for more details.
//
// You should have received a copy of the GNU Affero General Public License
// along with this program. If not, see <https://www.gnu.org/licenses/>.

package model

import (
	"bytes"
	"encoding/binary"
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

func Int64FromBytes(data []byte) (int64, []byte) {
	var result int64
	if len(data) < 8 {
		return 0, data
	}

	toRead := data[:8]
	remainder := data[8:]

	err := binary.Read(bytes.NewReader(toRead), binary.BigEndian, &result)
	if err != nil {
		return 0, data
	}

	return result, remainder
}

func Int64ToBytes(value int64) []byte {
	buf := new(bytes.Buffer)
	err := binary.Write(buf, binary.BigEndian, value)
	if err != nil {
		// TODO: not sure what to do here, log I guess when I am able to do that
		return []byte{}
	}
	return buf.Bytes()
}

func BoolToBytes(value bool) []byte {
	result := []byte{1}
	if value {
		result[0] = 1
	} else {
		result[0] = 0
	}
	return result
}

func BoolFromBytes(value []byte) (bool, []byte) {
	return value[0] == 1, value[1:]
}

func AddType(id uint8, data []byte) []byte {
	buf := make([]byte, len(data)+1)
	buf[0] = id
	copy(buf[1:], data)
	return buf
}

func BlockToBytes(value Block) []byte {
	id := StringToBytes(string(value.Id))
	data := BytesToBytes(value.Data)
	result := bytes.Join([][]byte{id, data}, []byte{})

	return result
}

func BlockFromBytes(value []byte) (Block, []byte) {
	id, remainder := StringFromBytes(value)
	data, remainder := BytesFromBytes(remainder)

	return Block{
		Id:   BlockId(id),
		Data: data,
	}, remainder
}
