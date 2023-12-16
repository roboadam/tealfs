package test

import (
	"tealfs/pkg/mgr"
	"tealfs/pkg/model/events"
	"tealfs/pkg/tnet"
	"testing"
	"time"
)

func TestThreeNodes(t *testing.T) {
	i1 := NewInputs()
	i2 := NewInputs()
	i3 := NewInputs()
	m1 := StartedMgr(i1)
	m2 := StartedMgr(i2)
	m3 := StartedMgr(i3)

	time.Sleep(time.Second * 5)

	i1.ConnectTo(i2)
	i2.ConnectTo(i3)

	time.Sleep(time.Second * 5)

	n1 := m1.GetRemoteNodes()
	n2 := m2.GetRemoteNodes()
	n3 := m3.GetRemoteNodes()
	if n1.Len() != 2 {
		t.Errorf("one had %d", n1.Len())
	}
	if n2.Len() != 2 {
		t.Errorf("two had %d", n2.Len())
	}
	if n3.Len() != 2 {
		t.Errorf("three had %d", n3.Len())
	}
}

func StartedMgr(inputs *Inputs) *mgr.Mgr {
	m := mgr.New(inputs.UiEvents, inputs.Net)
	m.Start()
	return &m
}

type Inputs struct {
	UiEvents chan events.Ui
	Net      *tnet.TcpNet
}

func (i *Inputs) ConnectTo(i2 *Inputs) {
	i.UiEvents <- events.Ui{EventType: events.ConnectTo, Argument: i2.Net.GetBinding()}
}

func NewInputs() *Inputs {
	net := tnet.NewTcpNet("localhost:0")
	return &Inputs{
		UiEvents: make(chan events.Ui),
		Net:      net,
	}
}
