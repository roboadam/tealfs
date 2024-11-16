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

package main

import (
	"context"
	"net/http"
	"testing"
	"time"
)

func TestOneNodeCluster(t *testing.T) {
	webdavAddress := "localhost:8080"
	uiAddress := "localhost:8081"
	nodeAddress := "localhost:8082"
	storagePath := "tmp"
	webdavUrl := "http://" + webdavAddress
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	startTealFs(storagePath, webdavAddress, uiAddress, nodeAddress, ctx)
	time.Sleep(time.Second) // TODO have this wait on a message if possible

	http.NewRequestWithContext(ctx, http.MethodPut, webdavAddress + "/text.txt", )
}
