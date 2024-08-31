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
			Name:    "/",
			IsDir:   true,
			Chidren: map[string]webdav.File{},
		},
	}
	fs.Mkdir(context.Background(), "/test", os.FileMode(0700))
}
