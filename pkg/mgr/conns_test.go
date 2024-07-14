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
	//c, _, _, inConnectTo, outStatuses := newConnsTest()

}

func newConnsTest() (Conns, chan ConnsMgrStatus, chan ConnsMgrReceive, chan MgrConnsConnectTo, chan MgrConnsSend, MockConnectionProvider) {
	outStatuses := make(chan ConnsMgrStatus)
	outReceives := make(chan ConnsMgrReceive)
	inConnectTo := make(chan MgrConnsConnectTo)
	inSends := make(chan MgrConnsSend)
	provider := MockConnectionProvider{}
	c := NewConns(outStatuses, outReceives, inConnectTo, inSends, provider)
	return c, outStatuses, outReceives, inConnectTo, inSends, provider
}
