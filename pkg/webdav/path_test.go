package webdav_test

import (
	"tealfs/pkg/webdav"
	"testing"
)

func TestPath(t *testing.T) {
	root := "/"
	p, err := webdav.PathFromName(root)
	if err != nil {
		t.Error("error parsing root:", err)
		return
	}
	if len(p) != 0 {
		t.Errorf("root should have no path parts")
		return
	}
}
