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
	"tealfs/pkg/model"
	"tealfs/pkg/webdav"
	"testing"

	"github.com/google/uuid"
)

func TestConfirmLock(t *testing.T) {
	ls := webdav.NewNetLockSystem()
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	go consumeConfirm(ctx, ls.ConfirmChan)
	go consumeRelease(ctx, ls.ReleaseChan)
	t.Error("")
}

func consumeConfirm(ctx context.Context, confirms chan webdav.LockConfirmReqResp) {
	for {
		select {
		case <-ctx.Done():
			return
		case confirm := <-confirms:
			confirm.Resp <- model.LockConfirmResponse{
				Ok:        true,
				ReleaseId: model.LockReleaseId(uuid.New().String()),
			}
		}
	}
}

func consumeRelease(ctx context.Context, releases chan model.LockReleaseId) {
	for {
		select {
		case <-ctx.Done():
			return
		case <-releases:
		}
	}
}
