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
