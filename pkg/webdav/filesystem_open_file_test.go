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

package webdav_test

import (
	"context"
	"os"
	"tealfs/pkg/webdav"
	"testing"
)

func TestCreateFile(t *testing.T) {
	fs := webdav.FileSystem{}
	_, err := fs.OpenFile(context.Background(), "/hello-world.txt", os.O_RDWR|os.O_CREATE, 0666)

	if err != nil {
		t.Error("Error opening file", err)
	}
}
