// Copyright (C) 2024 Adam Hess
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

package webdav

import "net/http"

type WebdavOps interface {
	ListenAndServe(addr string) error
	HandleFunc(pattern string, handler func(http.ResponseWriter, *http.Request))
	Handle(pattern string, handler http.Handler)
}

type HttpWebdavOps struct{}

func (h *HttpWebdavOps) ListenAndServe(addr string) error {
	return http.ListenAndServe(addr, nil)
}

func (h *HttpWebdavOps) Handle(pattern string, handler http.Handler) {
	http.Handle(pattern, handler)
}

type MockWebdavOps struct {
}

func (m *MockWebdavOps) ListenAndServe(addr string) error {
	return nil
}

func (m *MockWebdavOps) Handle(pattern string, handler http.Handler) {
}
