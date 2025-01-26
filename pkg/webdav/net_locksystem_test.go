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
	ls := webdav.NewNetLockSystem("node1")
	ctx, cancel := context.WithCancel(context.Background())
	releaseId := model.LockMessageId(uuid.New().String())
	defer cancel()
	go consumeMessages(ctx, ls.Messages, gwebdav.LockDetails{}, releaseId)
	go consumeRelease(ctx, ls.Release, t, releaseId)

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
	ls := webdav.NewNetLockSystem("node1")
	ctx, cancel := context.WithCancel(context.Background())
	expected := gwebdav.LockDetails{
		Root:      "root1",
		Duration:  time.Hour,
		OwnerXML:  "<a></a>",
		ZeroDepth: true,
	}
	defer cancel()
	go consumeMessages(ctx, ls.Messages, expected)

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

func consumeMessages(ctx context.Context, lockMessages chan webdav.LockMessageReqResp, details gwebdav.LockDetails, releaseIds ...model.LockMessageId) {
	remainder := releaseIds
	for {
		select {
		case <-ctx.Done():
			return
		case lockMessage := <-lockMessages:
			switch msg := lockMessage.Req.(type) {
			case *model.LockConfirmRequest:
				if len(remainder) > 0 {
					lockMessage.Resp <- &model.LockConfirmResponse{
						Ok:        true,
						ReleaseId: remainder[0],
						Id:        msg.Id,
						Caller:    msg.Caller,
					}
					remainder = remainder[1:]
				} else {
					lockMessage.Resp <- &model.LockConfirmResponse{
						Ok:        false,
						Message:   "No release id",
						ReleaseId: "",
						Id:        msg.Id,
						Caller:    msg.Caller,
					}
				}
			case *model.LockUnlockRequest:
				lockMessage.Resp <- &model.LockUnlockResponse{
					Ok:     true,
					Id:     msg.Id,
					Caller: msg.Caller,
				}
			case *model.LockCreateRequest:
				lockMessage.Resp <- &model.LockCreateResponse{
					Ok:     true,
					Token:  model.LockToken(uuid.New().String()),
					Id:     msg.Id,
					Caller: msg.Caller,
				}
			case *model.LockRefreshRequest:
				lockMessage.Resp <- &model.LockRefreshResponse{
					Ok:      true,
					Details: details,
					Id:      msg.Id,
					Caller:  msg.Caller,
				}
			}
		}
	}
}

func consumeRelease(ctx context.Context, releases chan model.LockMessageId, t *testing.T, expected ...model.LockMessageId) {
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
