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

package hash

import (
	"crypto/sha256"
)

type Hash struct {
	Value []byte
}

func FromRaw(rawHash []byte) Hash {
	return Hash{Value: rawHash}
}

func ForData(data []byte) Hash {
	value := sha256.Sum256(data)
	return Hash{value[:]}
}
