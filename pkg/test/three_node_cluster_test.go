package test

import (
	"os"
	"tealfs/pkg/mgr"
	"tealfs/pkg/model/events"
	"tealfs/pkg/store"
	"tealfs/pkg/tnet"
	"testing"
	"time"
)

func TestThreeNodes(t *testing.T) {
	i1 := NewInputs()
	i2 := NewInputs()
	i3 := NewInputs()
	m1 := StartedMgr(i1, t)
	m2 := StartedMgr(i2, t)
	m3 := StartedMgr(i3, t)

	i1.ConnectTo(i2)
	i2.ConnectTo(i3)

	time.Sleep(time.Second * 2)

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

	i1.AddData([]byte{1, 2, 3})
}

func StartedMgr(inputs *Inputs, t *testing.T) *mgr.Mgr {
	dir := tmpDir()
	defer cleanDir(dir, t)
	m := mgr.New(inputs.UiEvents, inputs.Net, dir)
	m.Start()
	return &m
}

type Inputs struct {
	UiEvents chan events.Event
	Net      *tnet.TcpNet
}

func (i *Inputs) ConnectTo(i2 *Inputs) {
	i.UiEvents <- events.NewString(events.ConnectTo, i2.Net.GetBinding())
}

func (i *Inputs) AddData(data []byte) {
	i.UiEvents <- events.NewBytes(events.AddData, data)
}

func NewInputs() *Inputs {
	net := tnet.NewTcpNet("localhost:0")
	return &Inputs{
		UiEvents: make(chan events.Event, 1_000),
		Net:      net,
	}
}
func removeAll(dir string, t *testing.T) {
	err := os.RemoveAll(dir)
	if err != nil {
		t.Errorf("Error [%v] deleting temp dir [%v]", err, dir)
	}
}

func tmpDir() store.Path {
	tempDir, _ := os.MkdirTemp("", "*-test")
	return store.NewPath(tempDir)
}

func cleanDir(path store.Path, t *testing.T) {
	removeAll(path.String(), t)
}
