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

package model_test

import (
	"tealfs/pkg/model"
	"testing"
	"time"

	"golang.org/x/net/webdav"
)

func TestConfirmRequest(t *testing.T) {
	cr := model.LockConfirmRequest{
		Now:   time.Now(),
		Name0: "name0val",
		Name1: "name1val",
		Conditions: []webdav.Condition{
			{
				Not:   false,
				Token: "token1",
				ETag:  "etag1",
			},
			{
				Not:   true,
				Token: "token2",
				ETag:  "etag2",
			},
		},
	}
	serialized := cr.ToBytes()
	crDeserialized := model.ToLockConfirmRequest(serialized)

	if !cr.Equal(crDeserialized) {
		t.Error("Expected values to be equal")
	}
}
