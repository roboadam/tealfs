package disk

import "os"

type FileOps interface {
	ReadFile(name string) ([]byte, error)
	WriteFile(name string, data []byte) error
}

type DiskFileOps struct{}

func (d *DiskFileOps) ReadFile(name string) ([]byte, error) {
	return os.ReadFile(name)
}

func (d *DiskFileOps) WriteFile(name string, data []byte) error {
	return os.WriteFile(name, data, 0644)
}

type MockFileOps struct {
	ReadError   error
	ReadData    []byte
	WriteError  error
	ReadPath    string
	WritePath   string
	WrittenData []byte
}

func (m *MockFileOps) ReadFile(name string) ([]byte, error) {
	m.ReadPath = name
	return m.ReadData, m.ReadError
}

func (m *MockFileOps) WriteFile(name string, data []byte) error {
	m.WrittenData = make([]byte, len(data))
	copy(m.WrittenData, data)
	m.WritePath = name
	return m.WriteError
}
