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

	"github.com/spf13/afero"
)

type FileOps interface {
	ReadFile(name string) ([]byte, error)
	WriteFile(name string, data []byte) error
	ReadDir(name string) ([]os.DirEntry, error)
	MkdirAll(name string) error
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

func (d *DiskFileOps) MkdirAll(name string) error {
	return os.MkdirAll(name, os.ModeDir)
}

type MockFileOps struct {
	ReadError  error
	WriteError error
	WriteCount int
	mockFS     *afero.IOFS
}

func (m *MockFileOps) ReadFile(name string) ([]byte, error) {
	if m.mockFS == nil {
		m.mockFS = &afero.IOFS{Fs: afero.NewMemMapFs()}
	}
	if m.ReadError != nil {
		return nil, m.ReadError
	}
	return m.mockFS.ReadFile(name)
}

func (m *MockFileOps) WriteFile(name string, data []byte) error {
	if m.mockFS == nil {
		m.mockFS = &afero.IOFS{Fs: afero.NewMemMapFs()}
	}
	if m.WriteError != nil {
		return m.WriteError
	}
	return afero.WriteFile(m.mockFS.Fs, name, data, 0644)
}

func (m *MockFileOps) ReadDir(name string) ([]os.DirEntry, error) {
	if m.mockFS == nil {
		m.mockFS = &afero.IOFS{Fs: afero.NewMemMapFs()}
	}
	return m.mockFS.ReadDir(name)
}

func (m *MockFileOps) MkdirAll(name string) error {
	if m.mockFS == nil {
		m.mockFS = &afero.IOFS{Fs: afero.NewMemMapFs()}
	}
	return m.mockFS.MkdirAll(name, os.ModeDir)
}
