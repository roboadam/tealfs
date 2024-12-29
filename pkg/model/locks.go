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

package model

import (
	"bytes"
	"time"

	"golang.org/x/net/webdav"
)

type LockMessageId string

func (l LockMessageId) GetId() LockMessageId {
	return l
}

func (l LockMessageId) AsPayload() Payload {
	return &l
}

func (l *LockMessageId) ToBytes() []byte {
	id := StringToBytes(string(*l))
	return AddType(LockMessageIdType, id)
}

func (l *LockMessageId) Equal(p Payload) bool {
	if o, ok := p.(*LockMessageId); ok {
		return l != o
	}
	return false
}

func ToLockMessageId(data []byte) *LockMessageId {
	id, _ := StringFromBytes(data)
	lId := LockMessageId(id)
	return &lId
}

type LockToken string

type LockConfirmRequest struct {
	Now          time.Time
	Name0, Name1 string
	Conditions   []webdav.Condition
	Id           LockMessageId
	Caller       NodeId
}

func (l *LockConfirmRequest) GetId() LockMessageId {
	return l.Id
}

func (l *LockConfirmRequest) AsPayload() Payload {
	return l
}

func (l *LockConfirmRequest) ToBytes() []byte {
	now := Int64ToBytes(l.Now.UnixMicro())
	name0 := StringToBytes(l.Name0)
	name1 := StringToBytes(l.Name1)
	numConditions := IntToBytes(uint32(len(l.Conditions)))
	conditionBytes := []byte{}
	for _, condition := range l.Conditions {
		conditionBytes = append(conditionBytes, ConditionToBytes(&condition)...)
	}
	id := StringToBytes(string(l.Id))
	caller := StringToBytes(string(l.Caller))
	return AddType(LockConfirmRequestType, bytes.Join([][]byte{now, name0, name1, numConditions, conditionBytes, id, caller}, []byte{}))
}

func ConditionToBytes(condition *webdav.Condition) []byte {
	not := BoolToBytes(condition.Not)
	token := StringToBytes(condition.Token)
	etag := StringToBytes(condition.ETag)
	return bytes.Join([][]byte{not, token, etag}, []byte{})
}

func ToCondition(data []byte) (webdav.Condition, []byte) {
	not, remainder := BoolFromBytes(data)
	token, remainder := StringFromBytes(remainder)
	etag, remainder := StringFromBytes(remainder)
	return webdav.Condition{Not: not, Token: token, ETag: etag}, remainder
}

func ConditionEquals(condition1 *webdav.Condition, condition2 *webdav.Condition) bool {
	if condition1.Not != condition2.Not {
		return false
	}
	if condition1.ETag != condition2.ETag {
		return false
	}
	if condition1.Token != condition2.Token {
		return false
	}
	return true
}

func (l *LockConfirmRequest) Equal(p Payload) bool {
	if o, ok := p.(*LockConfirmRequest); ok {
		if l.Now.UnixMicro() != o.Now.UnixMicro() {
			return false
		}
		if l.Name0 != o.Name0 {
			return false
		}
		if l.Name1 != o.Name1 {
			return false
		}
		for i, c := range l.Conditions {
			if !ConditionEquals(&c, &l.Conditions[i]) {
				return false
			}
		}
		if l.Caller != o.Caller {
			return false
		}
		return l.Id == o.Id
	}
	return false
}

func ToLockConfirmRequest(data []byte) *LockConfirmRequest {
	now, remainder := Int64FromBytes(data)
	name0, remainder := StringFromBytes(remainder)
	name1, remainder := StringFromBytes(remainder)
	numConditions, remainder := IntFromBytes(remainder)
	conditions := []webdav.Condition{}
	for range numConditions {
		var condition webdav.Condition
		condition, remainder = ToCondition(remainder)
		conditions = append(conditions, condition)
	}
	id, remainder := StringFromBytes(remainder)
	caller, _ := StringFromBytes(remainder)
	return &LockConfirmRequest{
		Now:        time.UnixMicro(now),
		Name0:      name0,
		Name1:      name1,
		Conditions: conditions,
		Id:         LockMessageId(id),
		Caller:     NodeId(caller),
	}
}

type LockConfirmResponse struct {
	Ok        bool
	Message   string
	ReleaseId LockMessageId
	Id        LockMessageId
	Caller    NodeId
}

func (l *LockConfirmResponse) GetId() LockMessageId {
	return l.Id
}

func (l *LockConfirmResponse) AsPayload() Payload {
	return l
}

func (l *LockConfirmResponse) ToBytes() []byte {
	success := BoolToBytes(l.Ok)
	message := StringToBytes(l.Message)
	releaseId := StringToBytes(string(l.ReleaseId))
	id := StringToBytes(string(l.Id))
	caller := StringToBytes(string(l.Caller))
	return AddType(LockConfirmResponseType, bytes.Join([][]byte{success, message, releaseId, id, caller}, []byte{}))
}

func (l *LockConfirmResponse) Equal(p Payload) bool {
	if o, ok := p.(*LockConfirmResponse); ok {
		if l.Message != o.Message {
			return false
		}
		if l.Ok != o.Ok {
			return false
		}
		if l.ReleaseId != o.ReleaseId {
			return false
		}
		if l.Id != o.Id {
			return false
		}
		if l.Caller != o.Caller {
			return false
		}
		return true
	}
	return false
}

func ToLockConfirmResponse(data []byte) *LockConfirmResponse {
	success, remainder := BoolFromBytes(data)
	message, remainder := StringFromBytes(remainder)
	releaseId, remainder := StringFromBytes(remainder)
	id, remainder := StringFromBytes(remainder)
	caller, _ := StringFromBytes(remainder)
	return &LockConfirmResponse{
		Ok:        success,
		Message:   message,
		ReleaseId: LockMessageId(releaseId),
		Id:        LockMessageId(id),
		Caller:    NodeId(caller),
	}
}

type LockCreateRequest struct {
	Now     time.Time
	Details webdav.LockDetails
	Id      LockMessageId
	Caller  NodeId
}

func (l *LockCreateRequest) GetId() LockMessageId {
	return l.Id
}

func (l *LockCreateRequest) AsPayload() Payload {
	return l
}

func LockDetailsToBytes(l *webdav.LockDetails) []byte {
	zeroDepth := BoolToBytes(l.ZeroDepth)
	ownerXml := StringToBytes(l.OwnerXML)
	root := StringToBytes(l.Root)
	duration := Int64ToBytes(int64(l.Duration))
	return bytes.Join([][]byte{zeroDepth, ownerXml, root, duration}, []byte{})
}

func (l *LockCreateRequest) ToBytes() []byte {
	now := Int64ToBytes(l.Now.UnixMicro())
	details := LockDetailsToBytes(&l.Details)
	id := StringToBytes(string(l.Id))
	caller := StringToBytes(string(l.Caller))
	return AddType(LockCreateRequestType, bytes.Join([][]byte{now, details, id, caller}, []byte{}))
}

func LockDetailsEquals(l1 *webdav.LockDetails, l2 *webdav.LockDetails) bool {
	if l1 == nil && l2 == nil {
		return true
	}

	if l1.Duration.Milliseconds() != l2.Duration.Milliseconds() {
		return false
	}

	if l1.OwnerXML != l2.OwnerXML {
		return false
	}

	if l1.Root != l2.Root {
		return false
	}

	if l1.ZeroDepth != l2.ZeroDepth {
		return false
	}

	return true
}

func ToLockDetails(data []byte) (webdav.LockDetails, []byte) {
	zeroDepth, remainder := BoolFromBytes(data)
	ownerXml, remainder := StringFromBytes(remainder)
	root, remainder := StringFromBytes(remainder)
	duration, remainder := Int64FromBytes(remainder)
	return webdav.LockDetails{
		Root:      root,
		Duration:  time.Duration(duration),
		OwnerXML:  ownerXml,
		ZeroDepth: zeroDepth,
	}, remainder
}

func (l *LockCreateRequest) Equal(p Payload) bool {
	if o, ok := p.(*LockCreateRequest); ok {
		if l.Now.UnixMilli() != o.Now.UnixMilli() {
			return false
		}
		if !LockDetailsEquals(&l.Details, &o.Details) {
			return false
		}
		if l.Id != o.Id {
			return false
		}
		if l.Caller != o.Caller {
			return false
		}
		return true
	}
	return false
}

func ToLockCreateRequest(data []byte) *LockCreateRequest {
	now, remainder := Int64FromBytes(data)
	details, remainder := ToLockDetails(remainder)
	id, remainder := StringFromBytes(remainder)
	caller, _ := StringFromBytes(remainder)
	return &LockCreateRequest{
		Now:     time.UnixMicro(now),
		Details: details,
		Id:      LockMessageId(id),
		Caller:  NodeId(caller),
	}
}

type LockCreateResponse struct {
	Token   LockToken
	Ok      bool
	Message string
	Id      LockMessageId
	Caller  NodeId
}

func (l *LockCreateResponse) GetId() LockMessageId {
	return l.Id
}

func (l *LockCreateResponse) AsPayload() Payload {
	return l
}

func (l *LockCreateResponse) ToBytes() []byte {
	token := StringToBytes(string(l.Token))
	ok := BoolToBytes(l.Ok)
	message := StringToBytes(l.Message)
	id := StringToBytes(string(l.Id))
	caller := StringToBytes(string(l.Caller))
	return AddType(LockCreateResponseType, bytes.Join([][]byte{token, ok, message, id, caller}, []byte{}))
}

func (l *LockCreateResponse) Equal(p Payload) bool {
	if o, ok := p.(*LockCreateResponse); ok {
		if l.Token != o.Token {
			return false
		}
		if l.Ok != o.Ok {
			return false
		}
		if l.Message != o.Message {
			return false
		}
		if l.Caller != o.Caller {
			return false
		}
		if l.Id != o.Id {
			return false
		}
		return true
	}
	return false
}

func ToLockCreateResponse(data []byte) *LockCreateResponse {
	token, remainder := StringFromBytes(data)
	ok, remainder := BoolFromBytes(remainder)
	message, remainder := StringFromBytes(remainder)
	id, remainder := StringFromBytes(remainder)
	caller, _ := StringFromBytes(remainder)
	return &LockCreateResponse{
		Token:   LockToken(token),
		Ok:      ok,
		Message: message,
		Id:      LockMessageId(id),
		Caller:  NodeId(caller),
	}
}

type LockRefreshRequest struct {
	Now      time.Time
	Token    LockToken
	Duration time.Duration
	Id       LockMessageId
	Caller   NodeId
}

func (l *LockRefreshRequest) GetId() LockMessageId {
	return l.Id
}

func (l *LockRefreshRequest) AsPayload() Payload {
	return l
}

func (l *LockRefreshRequest) ToBytes() []byte {
	now := Int64ToBytes(l.Now.UnixMicro())
	token := StringToBytes(string(l.Token))
	duration := Int64ToBytes(int64(l.Duration))
	id := StringToBytes(string(l.Id))
	caller := StringToBytes(string(l.Caller))
	return AddType(LockRefreshRequestType, bytes.Join([][]byte{now, token, duration, id, caller}, []byte{}))
}

func (l *LockRefreshRequest) Equal(p Payload) bool {
	if o, ok := p.(*LockRefreshRequest); ok {
		if l.Now.UnixMilli() != o.Now.UnixMilli() {
			return false
		}
		if l.Token != o.Token {
			return false
		}
		if l.Duration != o.Duration {
			return false
		}
		if l.Caller != o.Caller {
			return false
		}
		if l.Id != o.Id {
			return false
		}
		return true
	}
	return false
}

func ToLockRefreshRequest(data []byte) *LockRefreshRequest {
	now, remainder := Int64FromBytes(data)
	token, remainder := StringFromBytes(remainder)
	duration, remainder := Int64FromBytes(remainder)
	id, remainder := StringFromBytes(remainder)
	caller, _ := StringFromBytes(remainder)
	return &LockRefreshRequest{
		Now:      time.UnixMicro(now),
		Token:    LockToken(token),
		Duration: time.Duration(duration),
		Id:       LockMessageId(id),
		Caller:   NodeId(caller),
	}
}

type LockRefreshResponse struct {
	Details webdav.LockDetails
	Ok      bool
	Message string
	Id      LockMessageId
	Caller  NodeId
}

func (l *LockRefreshResponse) GetId() LockMessageId {
	return l.Id
}

func (l *LockRefreshResponse) AsPayload() Payload {
	return l
}

func (l *LockRefreshResponse) ToBytes() []byte {
	details := LockDetailsToBytes(&l.Details)
	ok := BoolToBytes(l.Ok)
	message := StringToBytes(l.Message)
	id := StringToBytes(string(l.Id))
	caller := StringToBytes(string(l.Caller))
	return AddType(LockRefreshResponseType, bytes.Join([][]byte{details, ok, message, id, caller}, []byte{}))
}

func (l *LockRefreshResponse) Equal(p Payload) bool {
	if o, ok := p.(*LockRefreshResponse); ok {
		if !LockDetailsEquals(&l.Details, &o.Details) {
			return false
		}
		if l.Ok != o.Ok {
			return false
		}
		if l.Message != o.Message {
			return false
		}
		if l.Id != o.Id {
			return false
		}
		if l.Caller != o.Caller {
			return false
		}
		return true
	}
	return false
}

func ToLockRefreshResponse(data []byte) *LockRefreshResponse {
	details, remainder := ToLockDetails(data)
	ok, remainder := BoolFromBytes(remainder)
	message, remainder := StringFromBytes(remainder)
	id, remainder := StringFromBytes(remainder)
	caller, _ := StringFromBytes(remainder)
	return &LockRefreshResponse{
		Details: details,
		Ok:      ok,
		Message: message,
		Id:      LockMessageId(id),
		Caller:  NodeId(caller),
	}
}

type LockUnlockRequest struct {
	Now    time.Time
	Token  LockToken
	Id     LockMessageId
	Caller NodeId
}

func (l *LockUnlockRequest) GetId() LockMessageId {
	return l.Id
}

func (l *LockUnlockRequest) AsPayload() Payload {
	return l
}

func (l *LockUnlockRequest) ToBytes() []byte {
	now := Int64ToBytes(l.Now.UnixMicro())
	token := StringToBytes(string(l.Token))
	id := StringToBytes(string(l.Id))
	caller := StringToBytes(string(l.Caller))
	return AddType(LockUnlockRequestType, bytes.Join([][]byte{now, token, id, caller}, []byte{}))
}

func (l *LockUnlockRequest) Equal(p Payload) bool {
	if o, ok := p.(*LockUnlockRequest); ok {
		if l.Now.UnixMilli() != o.Now.UnixMilli() {
			return false
		}
		if l.Token != o.Token {
			return false
		}
		if l.Caller != o.Caller {
			return false
		}
		if l.Id != o.Id {
			return false
		}
		return true
	}
	return false
}

func ToLockUnlockRequest(data []byte) *LockUnlockRequest {
	now, remainder := Int64FromBytes(data)
	token, remainder := StringFromBytes(remainder)
	id, remainder := StringFromBytes(remainder)
	caller, _ := StringFromBytes(remainder)
	return &LockUnlockRequest{
		Now:    time.UnixMicro(now),
		Token:  LockToken(token),
		Id:     LockMessageId(id),
		Caller: NodeId(caller),
	}
}

type LockUnlockResponse struct {
	Ok      bool
	Message string
	Id      LockMessageId
	Caller  NodeId
}

func (l *LockUnlockResponse) GetId() LockMessageId {
	return l.Id
}

func (l *LockUnlockResponse) AsPayload() Payload {
	return l
}

func (l *LockUnlockResponse) ToBytes() []byte {
	ok := BoolToBytes(l.Ok)
	message := StringToBytes(l.Message)
	id := StringToBytes(string(l.Id))
	caller := StringToBytes(string(l.Caller))
	return AddType(LockUnlockResponseType, bytes.Join([][]byte{ok, message, id, caller}, []byte{}))
}

func (l *LockUnlockResponse) Equal(p Payload) bool {
	if o, ok := p.(*LockUnlockResponse); ok {
		if l.Ok != o.Ok {
			return false
		}
		if l.Message != o.Message {
			return false
		}
		if l.Caller != o.Caller {
			return false
		}
		if l.Id != o.Id {
			return false
		}
		return true
	}
	return false
}

func ToLockUnlockResponse(data []byte) *LockUnlockResponse {
	ok, remainder := BoolFromBytes(data)
	message, remainder := StringFromBytes(remainder)
	id, remainder := StringFromBytes(remainder)
	caller, _ := StringFromBytes(remainder)
	return &LockUnlockResponse{
		Ok:      ok,
		Message: message,
		Id:      LockMessageId(id),
		Caller:  NodeId(caller),
	}
}
