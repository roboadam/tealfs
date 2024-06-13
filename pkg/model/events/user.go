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

package events

type Event struct {
	EventType Type
	argument  []byte
	result    chan []byte
}

func (e *Event) GetString() string {
	return string(e.argument)
}

func (e *Event) GetBytes() []byte {
	return e.argument
}

func (e *Event) GetResult() chan []byte {
	return e.result
}

func NewString(typ Type, argument string) Event {
	return Event{EventType: typ, argument: []byte(argument)}
}

func NewBytes(typ Type, argument []byte) Event {
	return Event{EventType: typ, argument: argument}
}

func NewBytesWithResult(typ Type, argument []byte, result chan []byte) Event {
	e := NewBytes(typ, argument)
	e.result = result
	return e
}

type Type int

const (
	ConnectTo Type = iota
	AddData
	ReadData
)
