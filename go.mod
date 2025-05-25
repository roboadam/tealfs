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

module tealfs

go 1.23.0

toolchain go1.24.3

require github.com/google/uuid v1.6.0

require golang.org/x/net v0.35.0

require (
	github.com/sirupsen/logrus v1.9.3 // indirect
	github.com/spf13/afero v1.14.0 // indirect
	golang.org/x/sys v0.30.0 // indirect
	golang.org/x/text v0.23.0 // indirect
)
