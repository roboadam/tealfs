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

package tnet

import (
	"encoding/gob"
	"io"
	"tealfs/pkg/model"
)

type RawNet struct {
	decoder *gob.Decoder
	encoder *gob.Encoder
	conn    io.ReadWriteCloser
}

func NewRawNet(conn io.ReadWriteCloser) *RawNet {
	return &RawNet{
		decoder: gob.NewDecoder(conn),
		encoder: gob.NewEncoder(conn),
		conn:    conn,
	}
}

func (r *RawNet) Close() error {
	return r.conn.Close()
}

func (r *RawNet) ReadPayload() (model.Payload, error) {
	var payload model.Payload
	err := r.decoder.Decode(&payload)
	return payload, err
}

func (r *RawNet) SendPayload(payload *model.Payload) error {
	err := r.encoder.Encode(payload)
	if err != nil {
		return err
	}
	return nil
}
