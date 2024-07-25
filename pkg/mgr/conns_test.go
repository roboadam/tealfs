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

package mgr

import (
	"bytes"
	"tealfs/pkg/hash"
	"tealfs/pkg/proto"
	"tealfs/pkg/store"
	"testing"
)

func TestAcceptConn(t *testing.T) {
	_, status, _, _, _, provider := newConnsTest()
	provider.Listener.accept <- true
	s := <-status
	if s.Type != Connected {
		t.Error("Received address")
	}
}

func TestConnectToConns(t *testing.T) {
	_, outStatus, _, inConnectTo, _, _ := newConnsTest()
	const expectedAddress = "expectedAddress:1234"
	status := connectTo(expectedAddress, outStatus, inConnectTo)
	if status.Type != Connected || status.RemoteAddress != expectedAddress {
		t.Error("Connection didn't work")
	}
}

func TestSendData(t *testing.T) {
	_, outStatus, _, inConnectTo, inSend, provider := newConnsTest()
	status := connectTo("address:123", outStatus, inConnectTo)
	inSend <- MgrConnsSend{
		ConnId: status.Id,
		Payload: &proto.SaveData{
			Block: store.Block{
				Id:   "blockId",
				Data: []byte{1, 2, 3},
				Hash: hash.ForData([]byte{1, 2, 3}),
			},
		}}

	// Fixme: Conn is null here
	writtenData := <-provider.Conn.dataWritten
	if !bytes.Equal(writtenData, []byte{3, 1, 2, 3}) {
		t.Error("Wrong data written")
	}
}

func connectTo(address string, outStatus chan ConnsMgrStatus, inConnectTo chan MgrConnsConnectTo) ConnsMgrStatus {
	inConnectTo <- MgrConnsConnectTo{address}
	return <-outStatus
}

func newConnsTest() (Conns, chan ConnsMgrStatus, chan ConnsMgrReceive, chan MgrConnsConnectTo, chan MgrConnsSend, MockConnectionProvider) {
	outStatuses := make(chan ConnsMgrStatus)
	outReceives := make(chan ConnsMgrReceive)
	inConnectTo := make(chan MgrConnsConnectTo)
	inSends := make(chan MgrConnsSend)
	provider := NewMockConnectionProvider()
	c := NewConns(outStatuses, outReceives, inConnectTo, inSends, &provider)
	return c, outStatuses, outReceives, inConnectTo, inSends, provider
}
