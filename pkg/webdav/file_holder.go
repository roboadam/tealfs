package webdav

import (
	"errors"
	"strings"
)

type pathSeg string
type path []pathSeg
type pathValue string
type fileHolder struct {
	data map[pathValue]File
}

func (f *fileHolder) add(file File) {
	f.data[file.path.toName()] = file
}

func (f *fileHolder) exists(p path) bool {
	_, exists := f.data[p.toName()]
	return exists
}

func newPathSeg(name string) (pathSeg, error) {
	if name == "" {
		return "", errors.New("invalid path segment")
	}
	if strings.Contains(name, "/") {
		return "", errors.New("invalid path segment")
	}
	return pathSeg(name), nil
}

func pathFromName(name string) (path, error) {
	if len(name) == 0 {
		return []pathSeg{}, errors.New("invalid path name, no leading slash")
	}
	return stringsToPath(strings.Split(name, "/")), nil
}

func (p path) toName() pathValue {
	name := ""
	for seg := range p {
		name += string(seg)
	}
	return pathValue(name)
}

func (p path) base() (path, error) {
	if len(p) == 0 {
		return nil, errors.New("empty path")
	}
	if len(p) == 1 {
		return []pathSeg{}, nil
	}
	return p[:len(p)-1], nil
}

func (p path) head() (pathSeg, error) {
	if len(p) == 0 {
		return "", errors.New("empty path")
	}
	return p[len(p)-1], nil
}

func (p path) startsWith(other path) bool {
	if len(other) > len(p) {
		return false
	}

	for i, otherPart := range other {
		if p[i] != otherPart {
			return false
		}
	}

	return true
}

func stringsToPath(strings []string) path {
	p := make(path, len(strings))
	for i, s := range strings {
		p[i] = pathSeg(s)
	}
	return p
}
