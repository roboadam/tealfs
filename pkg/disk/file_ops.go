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
	"os"
	"sync"
)

type FileOps interface {
	ReadFile(name string) ([]byte, error)
	WriteFile(name string, data []byte) error
	ListFiles(path string) ([]string, error)
}

type DiskFileOps struct{}

func (d *DiskFileOps) ReadFile(name string) ([]byte, error) {
	return os.ReadFile(name)
}

func (d *DiskFileOps) WriteFile(name string, data []byte) error {
	return os.WriteFile(name, data, 0644)
}

func (d *DiskFileOps) ListFiles(path string) ([]string, error) {
	result := []string{}
	files, err := os.ReadDir(path)
	if err != nil {
		return nil, err
	}
	for _, file := range files {
		if !file.IsDir() {
			result = append(result, file.Name())
		}
	}
	return result, nil
}

type MockFileOps struct {
	ReadError   error
	WriteError  error
	WriteCount  int
	mockFS      map[string][]byte
	mux         sync.Mutex
	DataWritten chan struct{}
}

func (m *MockFileOps) ReadFile(name string) ([]byte, error) {
	m.mux.Lock()
	defer m.mux.Unlock()
	if m.ReadError != nil {
		return nil, m.ReadError
	}
	if m.mockFS == nil {
		m.mockFS = make(map[string][]byte)
	}
	if data, ok := m.mockFS[name]; ok {
		return data, nil
	}
	return nil, os.ErrNotExist
}

func (m *MockFileOps) WriteFile(name string, data []byte) error {
	m.mux.Lock()
	defer m.mux.Unlock()
	if m.WriteError != nil {
		return m.WriteError
	}
	if m.mockFS == nil {
		m.mockFS = make(map[string][]byte)
	}
	m.mockFS[name] = data
	m.WriteCount++
	if m.DataWritten != nil {
		m.DataWritten <- struct{}{}
	}
	return nil
}

func (m *MockFileOps) ListFiles(path string) ([]string, error) {
	m.mux.Lock()
	defer m.mux.Unlock()
	if m.mockFS == nil {
		m.mockFS = make(map[string][]byte)
	}
	result := []string{}
	for name := range m.mockFS {
		result = append(result, name)
	}
	return result, nil
}
