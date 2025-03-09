package webdav

import (
	"errors"
	"strings"
	"tealfs/pkg/model"
	"tealfs/pkg/set"
)

type pathSeg string
type Path []pathSeg
type pathValue string
type FileHolder struct {
	byPath    map[pathValue]*File
	byBlockId map[model.BlockId]*File
}

func NewFileHolder() FileHolder {
	return FileHolder{
		byPath:    make(map[pathValue]*File),
		byBlockId: make(map[model.BlockId]*File),
	}
}

func (f *FileHolder) AllFiles() []*File {
	result := []*File{}
	for _, value := range f.byPath {
		result = append(result, value)
	}
	return result
}

func (f *FileHolder) Upsert(file *File) {
	
}

func (f *FileHolder) Add(file *File) {
	f.byPath[file.Path.toName()] = file
	f.byBlockId[file.Block.Id] = file
}

func (f *FileHolder) Delete(file *File) {
	delete(f.byPath, file.Path.toName())
	delete(f.byBlockId, file.Block.Id)
}

func (f *FileHolder) Exists(p Path) bool {
	_, exists := f.byPath[p.toName()]
	return exists
}

func (f *FileHolder) Get(p Path) (*File, bool) {
	file, exists := f.byPath[p.toName()]
	return file, exists
}

func (f *FileHolder) ToBytes() []byte {
	result := []byte{}
	for _, file := range f.byPath {
		result = append(result, file.ToBytes()...)
	}
	return result
}

func (f *FileHolder) updateFile(update *File) {
	toUpdate, exists := f.byBlockId[update.Block.Id]
	if exists {
		toUpdate.SizeValue = update.SizeValue
		toUpdate.ModeValue = update.ModeValue
		toUpdate.Modtime = update.Modtime
		oldPath := toUpdate.Path
		toUpdate.Path = update.Path
		delete(f.byPath, oldPath.toName())
		f.byPath[toUpdate.Path.toName()] = toUpdate
	} else {
		f.Add(update)
	}
}

func (f *FileHolder) UpdateFileHolderFromBytes(data []byte, fileSystem *FileSystem) error {
	if len(data) == 0 {
		return nil
	}
	allBlockIds := set.NewSet[model.BlockId]()
	remainderOverall := data
	for len(remainderOverall) > 0 {
		file, remainderFromFile, err := FileFromBytes(remainderOverall, fileSystem)
		remainderOverall = remainderFromFile
		if err != nil {
			return err
		}
		allBlockIds.Add(file.Block.Id)
		f.updateFile(&file)
	}
	for _, file := range f.AllFiles() {
		if !allBlockIds.Contains(file.Block.Id) {
			f.Delete(file)
		}
	}
	return nil
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
	return stringsToPath(stripEmptyStringsFromEnds(strings.Split(name, "/")))
}

func stripEmptyStringsFromEnds(values []string) []string {
	switch len(values) {
	case 0:
		return values
	case 1:
		if values[0] == "" {
			return []string{}
		}
		return values
	default:
		if values[0] == "" && values[len(values)-1] == "" {
			return values[1 : len(values)-1]
		}
		if values[0] == "" {
			return values[1:]
		}
		if values[len(values)-1] == "" {
			return values[:len(values)-1]
		}
		return values
	}
}

func (p Path) toName() pathValue {
	return pathValue(strings.Join(pathToStrings(p), "/"))
}

func (p Path) Equals(p2 Path) bool {
	if len(p) != len(p2) {
		return false
	}
	for i := range p {
		if p[i] != p2[i] {
			return false
		}
	}
	return true
}

func pathToStrings(input Path) []string {
	output := make([]string, len(input))
	for i, v := range input {
		output[i] = string(v) // Convert each element
	}
	return output
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

	return append(newPrefix, p[len(oldPrefix):]...)
}

func stringsToPath(strings []string) (Path, error) {
	p := make(Path, 0, len(strings))
	for _, s := range strings {
		seg, err := newPathSeg(s)
		if err != nil {
			return Path{}, err
		}
		p = append(p, seg)
	}
	return p, nil
}
