package tnet

import (
	"tealfs/pkg/proto"
	"tealfs/pkg/test"
	"testing"
)

func TestSaveData(t *testing.T) {
	//test_net := test.MockNet{
	//	AcceptsConnections: false,
	//	Conn:               test.Conn{},
	//}

	testConn := test.Conn{
		BytesWritten: make([]byte, 0),
	}

	dataToSave := []byte{0x01, 0x02, 0x03}
	payload := proto.SaveData{Data: dataToSave}
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
		if !equalSlices(p.Data, dataToSave) {
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
