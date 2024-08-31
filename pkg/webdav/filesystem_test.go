package webdav_test

import (
	"tealfs/pkg/webdav"
	"testing"
)

func TestMkdir(t *testing.T) {
	_ = webdav.FileSystem{
		Root: webdav.File{
			Name:    "/",
			IsDir:   true,
			Chidren: map[string]webdav.File{},
		},
	}
}
