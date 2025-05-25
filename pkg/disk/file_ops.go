// Copyright (C) 2025 Adam Hess
//
// This program is free software: you can redistribute it and/or modify it under
// the terms of the GNU Affero General Public License as published by the Free
// Software Foundation, version 3.
//
// This program is distributed in the hope that it will be useful, but WITHOUT
// ANY WARRANTY; without even the implied warranty of MERCHANTABILITY or
// FITNESS FOR A PARTICULAR PURPOSE. See the GNU Affero General Public License
// for more details.
//
// You should have received a copy of the GNU Affero General Public License
// along with this program. If not, see <https://www.gnu.org/licenses/>.

package disk

import (
	"io"
	"os"
	"sync"

	"github.com/spf13/afero"
)

type FileOps interface {
	ReadFile(name string) ([]byte, error)
	WriteFile(name string, data []byte) error
	ReadDir(name string) ([]os.DirEntry, error)
	CreateDir(name string) error
}

type DiskFileOps struct{}

func (d *DiskFileOps) ReadFile(name string) ([]byte, error) {
	return os.ReadFile(name)
}

func (d *DiskFileOps) WriteFile(name string, data []byte) error {
	return os.WriteFile(name, data, 0644)
}

func (d *DiskFileOps) ReadDir(name string) ([]os.DirEntry, error) {
	return os.ReadDir(name)
}

func (d *DiskFileOps) CreateDir(name string) error {
	return os.MkdirAll(name, os.ModeDir)
}

type MockFileOps struct {
	ReadError  error
	WriteError error
	WriteCount int
	mockFS     afero.Fs
	mux        sync.Mutex
}

func (m *MockFileOps) ReadFile(name string) ([]byte, error) {
	if m.mockFS == nil {
		m.mockFS = afero.NewMemMapFs()
	}
	if m.ReadError != nil {
		return nil, m.ReadError
	}
	f, err := m.mockFS.OpenFile(name, os.O_RDONLY, 0644)
	if err != nil {
		return []byte{}, err
	}
	data, err := io.ReadAll(f)
	f.Close()
	return data, err
}

func (m *MockFileOps) WriteFile(name string, data []byte) error {
	m.mux.Lock()
	defer m.mux.Unlock()
	if m.WriteError != nil {
		return m.WriteError
	}

	f, err := m.mockFS.OpenFile(name, os.O_WRONLY, 0666)
	if err != nil {
		return err
	}
	_, err = f.Write([]byte(name))
	f.Close()
	if err == nil {
		m.WriteCount++
	}
	return err
}

func (d *MockFileOps) ReadDir(name string) ([]os.DirEntry, error) {
	return nil, os.ErrNotExist
}

func (d *MockFileOps) CreateDir(name string) error {
	d.mockFS.ReadDir(``)
	return
}
