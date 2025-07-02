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

package conns

import (
	"encoding/gob"
	"tealfs/pkg/model"
)


func (c *Conns) sendPayload(conn model.ConnId, payload model.Payload) error {
	encoder := gob.NewEncoder(c.netConns[conn])
	err := encoder.Encode(payload.Type())
	if err != nil {
		return err
	}
	err = encoder.Encode(payload)
	if err != nil {
		return err
	}
	return nil
}

// func intForType(val any) uint16 {
// 	switch reflect.TypeOf(val) {
// 	case reflect.TypeOf(model.IAm{}):
// 		{
// 			return 0
// 		}
// 	case reflect.TypeOf(model.WriteRequest{}):
// 		{
// 			return 1
// 		}
// 	case reflect.TypeOf(model.ReadRequest{}):
// 		{
// 			return 2
// 		}
// 	default:
// 		{
// 			panic("unknown type")
// 		}
// 	}
// }
