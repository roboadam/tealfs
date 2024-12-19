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
	"time"

	"github.com/google/uuid"
	gwebdav "golang.org/x/net/webdav"
)

func TestConfirmLock(t *testing.T) {
	ls := webdav.NewNetLockSystem()
	ctx, cancel := context.WithCancel(context.Background())
	releaseId := model.LockReleaseId(uuid.New().String())
	defer cancel()
	go consumeConfirm(ctx, ls.ConfirmChan, releaseId)
	go consumeRelease(ctx, ls.ReleaseChan, t, releaseId)
	go consumeCreate(ctx, ls.CreateChan)
	go consumeUnlock(ctx, ls.UnlockChan)

	err := ls.Unlock(time.Now(), "token1")
	if err != nil {
		t.Error("Error Unlocking", err)
	}

	condition1 := gwebdav.Condition{
		Not:   true,
		Token: "cToken1",
		ETag:  "etag1",
	}
	release, err := ls.Confirm(time.Now(), "name0", "name1", condition1)
	if err != nil {
		t.Error("Error Unlocking", err)
	}
	release()

	token, err := ls.Create(time.Now(), gwebdav.LockDetails{
		Root:      "root2",
		Duration:  time.Hour * 2,
		OwnerXML:  "<p></p>",
		ZeroDepth: true,
	})
	if err != nil {
		t.Error("Error Creating", err)
	}
	if len(token) == 0 {
		t.Error("Empty token")
	}
}

func TestRefreshLock(t *testing.T) {
	ls := webdav.NewNetLockSystem()
	ctx, cancel := context.WithCancel(context.Background())
	expected := gwebdav.LockDetails{
		Root:      "root1",
		Duration:  time.Hour,
		OwnerXML:  "<a></a>",
		ZeroDepth: true,
	}
	defer cancel()
	go consumeRefresh(ctx, ls.RefreshChan, expected)

	received, err := ls.Refresh(time.Now(), "token1", time.Hour)
	if err != nil {
		t.Error("Error Refreshing", err)
		return
	}
	if received != expected {
		t.Error("Unexpected result")
		return
	}
}

func consumeConfirm(ctx context.Context, confirms chan webdav.LockConfirmReqResp, releaseIds ...model.LockReleaseId) {
	remainder := releaseIds
	for {
		select {
		case <-ctx.Done():
			return
		case confirm := <-confirms:
			if len(remainder) > 0 {
				confirm.Resp <- &model.LockConfirmResponse{
					Ok:        true,
					ReleaseId: remainder[0],
				}
				remainder = remainder[1:]
			} else {
				confirm.Resp <- &model.LockConfirmResponse{
					Ok:        false,
					ReleaseId: "",
				}
			}
		}
	}
}

func consumeRelease(ctx context.Context, releases chan model.LockReleaseId, t *testing.T, expected ...model.LockReleaseId) {
	remainder := expected
	for {
		select {
		case <-ctx.Done():
			return
		case received := <-releases:
			if len(remainder) > 0 {
				if remainder[0] != received {
					t.Error("Unexpected release id")
				}
				remainder = remainder[1:]
			}
		}
	}
}

func consumeCreate(ctx context.Context, creates chan webdav.LockCreateReqResp) {
	for {
		select {
		case <-ctx.Done():
			return
		case create := <-creates:
			create.Resp <- model.LockCreateResponse{
				Token: model.LockToken(uuid.New().String()),
				Ok:    true,
			}
		}
	}
}

func consumeRefresh(ctx context.Context, refreshes chan webdav.LockRefreshReqResp, details gwebdav.LockDetails) {
	for {
		select {
		case <-ctx.Done():
			return
		case refresh := <-refreshes:
			refresh.Resp <- model.LockRefreshResponse{
				Details: details,
				Ok:      true,
			}
		}
	}
}

func consumeUnlock(ctx context.Context, unlocks chan webdav.LockUnlockReqResp) {
	for {
		select {
		case <-ctx.Done():
			return
		case unlock := <-unlocks:
			unlock.Resp <- model.LockUnlockResponse{
				Ok: true,
			}
		}
	}
}
