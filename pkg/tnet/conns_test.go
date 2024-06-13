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

package tnet

import (
	"tealfs/pkg/proto"
	"tealfs/pkg/store"
	"tealfs/pkg/test"
	"testing"
)

func TestSaveData(t *testing.T) {
	testConn := test.Conn{
		BytesWritten: make([]byte, 0),
	}

	dataToSave := []byte{0x01, 0x02, 0x03}
	payload := proto.SaveData{Block: store.Block{Data: dataToSave}}
	err := SendPayload(&testConn, payload.ToBytes())
	if err != nil {
		t.Error("Error!")
	}

	testConn2 := test.Conn{
		BytesToRead: testConn.BytesWritten,
	}

	var payloadBytes []byte
	payloadBytes, err = ReadPayload(&testConn2)

	samePayload := proto.ToPayload(payloadBytes)

	switch p := samePayload.(type) {
	case *proto.SaveData:
		if !equalSlices(p.Block.Data, dataToSave) {
			t.Error("What data??")
		}
	default:
		t.Error("What type??")
	}
}

func equalSlices(b1 []byte, b2 []byte) bool {
	if len(b1) != len(b2) {
		return false
	}

	for i, b := range b1 {
		if b != b2[i] {
			return false
		}
	}

	return true
}
