package webdav_test

import (
	"context"
	"os"
	"tealfs/pkg/webdav"
	"testing"
)

func TestMkdir(t *testing.T) {
	fs := webdav.FileSystem{
		Root: webdav.File{
			NameValue:    "/",
			IsDirValue:   true,
			Chidren: map[string]webdav.File{},
		},
	}
	c := context.Background()
	mode := os.FileMode(0700)

	err := fs.Mkdir(c, "/test", mode)
	if err != nil || !dirOpenedOk(fs, "/test") {
		t.Error("can't open dir")
	}

	err = fs.Mkdir(c, "/test/stuff", mode)
	if err != nil || !dirOpenedOk(fs, "/test") {
		t.Error("can't open dir")
	}

	err = fs.Mkdir(c, "/new/stuff", mode)
	if err == nil {
		t.Error("shouldn't have been able to open this one")
	}
}

func dirOpenedOk(fs webdav.FileSystem, name string) bool {
	f, err := fs.OpenFile(context.Background(), name, os.O_RDONLY, 0600)
	if err != nil {
		return false
	}
	if f == nil {
		return false
	}
	return true
}
