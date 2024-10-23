package webdav_test

import (
	"tealfs/pkg/webdav"
	"testing"
)

func TestValidPaths(t *testing.T) {
	p, err := webdav.PathFromName("/")
	expectResultOfLen(t, p, err, 0)

	p, err = webdav.PathFromName("/asdf")
	expectResultOfLen(t, p, err, 1)

	p, err = webdav.PathFromName("/qwer/ /qwer")
	expectResultOfLen(t, p, err, 3)

	p, err = webdav.PathFromName("/qwer/qwer/")
	expectResultOfLen(t, p, err, 2)

	p, err = webdav.PathFromName("/qwer/asdf/lkj")
	expectResultOfLen(t, p, err, 3)
}

func TestInValidPaths(t *testing.T) {
	_, err := webdav.PathFromName("")
	expectError(t, err)

	_, err = webdav.PathFromName("asdf")
	expectError(t, err)

	_, err = webdav.PathFromName("asdf/asdf")
	expectError(t, err)

	_, err = webdav.PathFromName("/adf//asdf")
	expectError(t, err)

}

func expectError(t *testing.T, err error) {
	if err == nil {
		t.Errorf("error expected")
	}
}

func expectResultOfLen(t *testing.T, path webdav.Path, err error, count int) {
	if err != nil {
		t.Error("error parsing path:", err)
		return
	}
	if len(path) != count {
		t.Errorf("unexpected number of path parts")
		return
	}
}
