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

package conns

import (
	"bytes"
	"encoding/binary"
	"tealfs/pkg/hash"
	"tealfs/pkg/model"
	"testing"
)

func TestAcceptConn(t *testing.T) {
	_, status, _, _, _, provider := newConnsTest()
	provider.Listener.accept <- true
	s := <-status
	if s.Type != model.Connected {
		t.Error("Received address")
	}
}

func TestConnectToConns(t *testing.T) {
	_, outStatus, _, inConnectTo, _, _ := newConnsTest()
	const expectedAddress = "expectedAddress:1234"
	status := connectTo(expectedAddress, outStatus, inConnectTo)
	if status.Type != model.Connected || status.RemoteAddress != expectedAddress {
		t.Error("Connection didn't work")
	}
}

func TestSendData(t *testing.T) {
	_, outStatus, _, inConnectTo, inSend, provider := newConnsTest()
	status := connectTo("address:123", outStatus, inConnectTo)
	inSend <- model.MgrConnsSend{
		ConnId: status.Id,
		Payload: &model.SaveData{
			Block: model.Block{
				Id:   "blockId",
				Data: []byte{1, 2, 3},
				Hash: hash.ForData([]byte{1, 2, 3}),
			},
		}}

	expectedBytes := []byte{0, 0, 0, 55,
		3, 0, 0, 0, 7, 98, 108, 111, 99, 107, 73, 100, 0, 0, 0, 3, 1, 2, 3, 0, 0, 0, 32, 3,
		144, 88, 198, 242, 192, 203, 73, 44, 83, 59, 10, 77, 20, 239, 119, 204, 15, 120, 171,
		204, 206, 213, 40, 125, 132, 161, 162, 1, 28, 251, 129}
	if !dataMatched(expectedBytes, provider.Conn.dataWritten) {
		t.Error("Wrong data written")
	}
}

func dataMatched(expected []byte, incoming chan []byte) bool {
	buffer := make([]byte, 0)
	for readBytes := range incoming {
		buffer = append(buffer, readBytes...)
		if len(buffer) >= len(expected) {
			return bytes.Equal(expected, buffer[:len(expected)])
		}
	}
	return false
}

func TestGetData(t *testing.T) {
	_, outStatus, cmr, inConnectTo, _, provider := newConnsTest()
	status := connectTo("address:123", outStatus, inConnectTo)
	payload := &model.IAm{
		NodeId: "nodeId",
	}
	dataReceived := payload.ToBytes()
	length := lenAsBytes(dataReceived)
	provider.Conn.dataToRead <- length
	provider.Conn.dataToRead <- dataReceived

	result := <-cmr

	if result.ConnId != status.Id || !result.Payload.Equal(payload) {
		t.Error("We didn't pass the message")
	}
}

func lenAsBytes(data []byte) []byte {
	size := uint32(len(data))
	buf := make([]byte, 4)
	binary.BigEndian.PutUint32(buf, size)
	return buf
}

func connectTo(address string, outStatus chan model.ConnsMgrStatus, inConnectTo chan model.MgrConnsConnectTo) model.ConnsMgrStatus {
	inConnectTo <- model.MgrConnsConnectTo{Address: address}
	return <-outStatus
}

func newConnsTest() (Conns, chan model.ConnsMgrStatus, chan model.ConnsMgrReceive, chan model.MgrConnsConnectTo, chan model.MgrConnsSend, MockConnectionProvider) {
	outStatuses := make(chan model.ConnsMgrStatus)
	outReceives := make(chan model.ConnsMgrReceive)
	inConnectTo := make(chan model.MgrConnsConnectTo)
	inSends := make(chan model.MgrConnsSend)
	provider := NewMockConnectionProvider()
	c := NewConns(outStatuses, outReceives, inConnectTo, inSends, &provider)
	return c, outStatuses, outReceives, inConnectTo, inSends, provider
}
