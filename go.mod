// Copyright (C) 2026 Adam Hess
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

module tealfs

go 1.23.0

toolchain go1.24.4

require github.com/google/uuid v1.6.0

require (
	github.com/sirupsen/logrus v1.9.3
	golang.org/x/net v0.41.0
)

require golang.org/x/sys v0.33.0 // indirect
