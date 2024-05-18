package mgr

import (
	"testing"
)

func TestConnectToRemoteNodeNew(t *testing.T) {
	const expectedAddress = "some-address:123"

	m := NewNew()
	m.Start()

	m.UiMgrConnectTos <- UiMgrConnectTo{
		Address: expectedAddress,
	}

	expectedMessage := <-m.MgrConnsConnectTos

	if expectedMessage.Address != expectedAddress {
		t.Error("Received address", expectedMessage.Address)
	}
}

func TestConnectToSuccess(t *testing.T) {
	const expectedAddress = "some-address:123"
	const expectedConnectionId = 1

	m := NewNew()
	m.Start()

	m.ConnsMgrStatuses <- ConnsMgrStatus{
		Type:          Connected,
		RemoteAddress: expectedAddress,
		Id:            expectedConnectionId,
	}

	// Fixme: Verify mgr sends IAM
	// Fixme: Understand what connAddress is for and test it or remove it
}
