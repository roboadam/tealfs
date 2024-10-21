package webdav

import (
	"errors"
	"strings"
)

type pathSeg string
type Path []pathSeg
type pathValue string
type fileHolder struct {
	data map[pathValue]File
}

func (f *fileHolder) allFiles() []File {
	result := []File{}
	for _, value := range f.data {
		result = append(result, value)
	}
	return result
}

func (f *fileHolder) add(file File) {
	f.data[file.path.toName()] = file
}

func (f *fileHolder) delete(path Path) {
	delete(f.data, path.toName())
}

func (f *fileHolder) exists(p Path) bool {
	_, exists := f.data[p.toName()]
	return exists
}

func (f *fileHolder) get(p Path) (File, bool) {
	file, exists := f.data[p.toName()]
	return file, exists
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

func PathFromName(name string) (Path, error) {
	if len(name) == 0 {
		return []pathSeg{}, errors.New("invalid path name, no leading slash")
	}
	return stringsToPath(strings.Split(name, "/"))
}

func (p Path) toName() pathValue {
	name := ""
	for _, seg := range p {
		name += string(seg)
	}
	return pathValue(name)
}

func (p Path) base() (Path, error) {
	if len(p) == 0 {
		return nil, errors.New("empty path")
	}
	if len(p) == 1 {
		return []pathSeg{}, nil
	}
	return p[:len(p)-1], nil
}

func (p Path) head() (pathSeg, error) {
	if len(p) == 0 {
		return "", errors.New("empty path")
	}
	return p[len(p)-1], nil
}

func (p Path) startsWith(other Path) bool {
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

func (p Path) swapPrefix(oldPrefix Path, newPrefix Path) Path {
	if !p.startsWith(oldPrefix) {
		return p
	}

	return append(newPrefix, p[:len(oldPrefix)]...)
}

func stringsToPath(strings []string) (Path, error) {
	p := make(Path, 0, len(strings)-1)
	for i, s := range strings {
		seg, err := newPathSeg(s)
		if err != nil {
			return Path{}, err
		}
		p[i] = seg
	}
	return p, nil
}
